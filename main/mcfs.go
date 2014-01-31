/*
 * This package implements the Materials Commons File Server service. This
 * service provides upload/download of datafiles from the Materials Commons
 * repository.

 * The protocol for file uploads looks as follows:
 *     1. The client sends the size, checksum and path. If the file
 *        is an existing file then it also sends the DataFileID for
 *        the file.
 *
 *     2. If the server receives a DataFileID it checks the size
 *        and checksum against what was sent. If the checksums
 *        match and the sizes are different then its a partially
 *        completed upload. If the checksums are different then
 *        its a new upload.
 *
 *     3. The server sends back the DataFileID. It will create a
 *        new DataFileID or send back an existing depending on
 *        whether its a new upload or an existing one.
 *
 *     4. The server will tell the client the offset to start
 *        sending data from. For a new upload this will be at
 *        position 0. For an existing one it will be the offset
 *        to restart the upload.
 *
 * The protocol for file downloads looks as follows:
 *
 */
package main

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/jessevdk/go-flags"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/contrib/schema"
	"github.com/materials-commons/materials/util"
	_ "github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

// Options for server startup
type ServerOptions struct {
	Port     uint   `long:"server-port" description:"The port the server listens on" default:"35862"`
	Bind     string `long:"bind" description:"Address of local interface to listen on" default:"localhost"`
	MCDir    string `long:"mcdir" description:"Directory path to materials commons file storage" default:"/mcfs/data/materialscommons"`
	PrintPid bool   `long:"print-pid" description:"Prints the server pid to stdout"`
	HttpPort uint   `long:"http-port" description:"Port webserver listens on" default:"5010"`
}

// Options for the database
type DatabaseOptions struct {
	Connection string `long:"db-connect" description:"The host/port to connect to database on" default:"localhost:28015"`
	Name       string `long:"db" description:"Database to use" default:"materialscommons"`
}

// Break the options into option groups.
type Options struct {
	Server   ServerOptions   `group:"Server Options"`
	Database DatabaseOptions `group:"Database Options"`
}

// The following are set to command line argument values
var MCDir string     // Directory datafiles are stored in
var DBAddress string // Database address
var DBName string    // Database name

func main() {
	var opts Options
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

	MCDir = opts.Server.MCDir
	DBAddress = opts.Database.Connection
	DBName = opts.Database.Name
	go webserver(opts.Server.HttpPort)

	acceptConnections(listener)
}

// webserver starts an http server that serves out datafile.
func webserver(port uint) {
	http.HandleFunc("/datafiles/static/", datafileHandler)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// datafileHandler server data files.
func datafileHandler(writer http.ResponseWriter, req *http.Request) {
	apikey := req.FormValue("apikey")
	if apikey == "" {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	session, err := r.Connect(map[string]interface{}{
		"address":  DBAddress,
		"database": DBName,
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
	df, err := model.GetDataFile(dataFileID, session)
	switch {
	case err != nil:
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case !request.OwnerGaveAccessTo(df.Owner, u.Email, session):
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	default:
		var idToUse string
		if df.UsesID != "" {
			idToUse = df.UsesID
		} else {
			idToUse = dataFileID
		}
		path := request.DataFilePath(MCDir, idToUse)
		http.ServeFile(writer, req, path)
	}
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
			"address":  DBAddress,
			"database": DBName,
		})
		if err != nil {
			conn.Close()
			continue
		}

		m := util.NewGobMarshaler(conn)
		r := request.NewReqHandler(m, session, MCDir)
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
