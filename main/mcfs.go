/*
This package implements the Materials Commons File Server service. This
service provides upload/download of datafiles from the Materials Commons
repository.

The protocol for file uploads looks as follows:
    1. The client sends the size, checksum and path. If the file
       is an existing file then it also sends the DataFileID for
       the file.

    2. If the server receives a DataFileID it checks the size
       and checksum against what was sent. If the checksums
       match and the sizes are different then its a partially
       completed upload. If the checksums are different then
       its a new upload.

    3. The server sends back the DataFileID. It will create a
       new DataFileID or send back an existing depending on
       whether its a new upload or an existing one.

    4. The server will tell the client the offset to start
       sending data from. For a new upload this will be at
       position 0. For an existing one it will be the offset
       to restart the upload.

The protocol for file downloads looks as follows:

*/
package main

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/jessevdk/go-flags"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/materials/util"
	_ "github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"mime"
)

// Options for server startup
type serverOptions struct {
	Port     uint   `long:"server-port" description:"The port the server listens on" default:"35862"`
	Bind     string `long:"bind" description:"Address of local interface to listen on" default:"localhost"`
	MCDir    string `long:"mcdir" description:"Directory path to materials commons file storage" default:"/mcfs/data/materialscommons"`
	PrintPid bool   `long:"print-pid" description:"Prints the server pid to stdout"`
	HTTPPort uint   `long:"http-port" description:"Port webserver listens on" default:"5010"`
}

// Options for the database
type databaseOptions struct {
	Connection string `long:"db-connect" description:"The host/port to connect to database on" default:"localhost:28015"`
	Name       string `long:"db" description:"Database to use" default:"materialscommons"`
}

// Break the options into option groups.
type options struct {
	Server   serverOptions   `group:"Server Options"`
	Database databaseOptions `group:"Database Options"`
}

// The following are set to command line argument values
var mcDir string     // Directory datafiles are stored in
var dbAddress string // Database address
var dbName string    // Database name

func main() {
	var opts options
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	listener, err := createListener(opts.Server.Bind, opts.Server.Port)
	if err != nil {
		os.Exit(1)
	}

	if opts.Server.PrintPid {
		fmt.Println(os.Getpid())
	}

	mcDir = opts.Server.MCDir
	dbAddress = opts.Database.Connection
	dbName = opts.Database.Name
	go webserver(opts.Server.HTTPPort)

	acceptConnections(listener)
}

// webserver starts an http server that serves out datafile.
func webserver(port uint) {
	http.HandleFunc("/datafiles/static/", datafileHandler)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// datafileHandler serves data files.
func datafileHandler(writer http.ResponseWriter, req *http.Request) {
	apikey := req.FormValue("apikey")
	if apikey == "" {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	download := req.FormValue("download")

	session, err := r.Connect(map[string]interface{}{
		"address":  dbAddress,
		"database": dbName,
	})

	if err != nil {
		http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	defer session.Close()

	// Verify key
	var u schema.User
	query := r.Table("users").GetAllByIndex("apikey", apikey)
	if err := model.GetRow(query, session, &u); err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get datafile from db and check access
	dataFileID := filepath.Base(req.URL.Path)
	df, err := model.GetFile(dataFileID, session)
	switch {
	case err != nil:
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case !request.OwnerGaveAccessTo(df.Owner, u.Email, session):
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	default:
		var path string
		if isTiff(df.Name) && download == "" {
			path = tiffConversionPath(mcDir, idToUse(df))
		} else {
			extension := strings.ToLower(filepath.Ext(df.Name))
			if mimetype := mime.TypeByExtension(extension); mimetype != "" {
				writer.Header().Set("Content-Type", mimetype)
			}
			path = request.DataFilePath(mcDir, idToUse(df))
		}
		http.ServeFile(writer, req, path)
	}
}

// isTiff checks a name to see if it is for a TIFF file.
func isTiff(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".tiff" || ext == ".tif" {
		return true
	}
	return false
}

func tiffConversionPath(mcdir, id string) string {
	return filepath.Join(request.DataFileDir(mcdir, id), ".conversion", id+".jpg")
}

// idToUse looks at a datafile and if UsesID is set returns that id
// otherwise it returns Id. UsesID is set when a file has already
// been uploaded and their is a duplicate entry. Duplicates point
// to the real file through UsesID.
func idToUse(dataFile *schema.File) string {
	if dataFile.UsesID != "" {
		return dataFile.UsesID
	}

	return dataFile.ID
}

// createListener creates the net connection. It connects to the specified host
// and port.
func createListener(host string, port uint) (*net.TCPListener, error) {
	service := fmt.Sprintf("%s:%d", host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		fmt.Println("Resolve error:", err)
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Listen error:", err)
		return nil, err
	}

	return listener, nil
}

// acceptConnections listens on the the TCPListener. When a new connection comes
// in it is dispatched in a separate go routine. For each new connection a new
// connection the to database is created.
func acceptConnections(listener *net.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		session, err := r.Connect(map[string]interface{}{
			"address":  dbAddress,
			"database": dbName,
		})
		if err != nil {
			conn.Close()
			continue
		}

		m := util.NewGobMarshaler(conn)
		r := request.NewReqHandler(m, session, mcDir)
		go handleConnection(r, conn, session)
	}
}

// handleConnection handles connection requests by running the state machine. It also
// takes care of book keeping like shutting down the net and database connections when
// the connection is terminated.
func handleConnection(reqHandler *request.ReqHandler, conn net.Conn, session *r.Session) {
	defer conn.Close()
	defer session.Close()

	reqHandler.Run()
}
