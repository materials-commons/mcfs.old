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
*/
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/jessevdk/go-flags"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcfs/base/db"
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/client/util"
	_ "github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/request"
	"github.com/materials-commons/mcfs/server/service"
)

// Options for server startup
type serverOptions struct {
	Port     uint   `long:"server-port" description:"The port the server listens on" default:"35862"`
	Bind     string `long:"bind" description:"Address of local interface to listen on" default:"localhost"`
	MCDir    string `long:"mcdir" description:"Directory path to materials commons file storage"`
	PrintPid bool   `long:"print-pid" description:"Prints the server pid to stdout"`
	HTTPPort uint   `long:"http-port" description:"Port webserver listens on" default:"5010"`
}

// Options for the database
type databaseOptions struct {
	Connection string `long:"db-connect" description:"The database connection string"`
	Name       string `long:"db" description:"Database to use"`
	Type       string `long:"db-type" description:"The type of database to connect to"`
}

// Break the options into option groups.
type options struct {
	Server   serverOptions   `group:"Server Options"`
	Database databaseOptions `group:"Database Options"`
}

func configErrorHandler(key string, err error, args ...interface{}) {

}

var s *service.Service

func setupRethinkDB() {
	dbConn := config.GetString("MCDB_CONNECTION")
	dbName := config.GetString("MCDB_NAME")
	db.SetAddress(dbConn)
	db.SetDatabase(dbName)
}

func init() {
	config.Init(config.TwelveFactorWithOverride)
	config.SetErrorHandler(configErrorHandler)
	setupRethinkDB()
	s = service.New(service.RethinkDB)
}

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

	setupConfig(opts.Database, opts.Server)

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s: %s\n", e, debug.Stack())
		}
	}()

	go webserver(opts.Server.HTTPPort)

	acceptConnections(listener)
}

func setupConfig(dbOpts databaseOptions, serverOpts serverOptions) {
	if dbOpts.Connection != "" {
		config.Set("MCDB_CONNECTION", dbOpts.Connection)
	}

	if dbOpts.Name != "" {
		config.Set("MCDB_NAME", dbOpts.Name)
	}

	if dbOpts.Type != "" {
		config.Set("MCDB_TYPE", dbOpts.Type)
	}

	if serverOpts.MCDir != "" {
		config.Set("MCDIR", serverOpts.MCDir)
	}
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

	// Verify key
	u, err := s.User.ByAPIKey(apikey)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get datafile from db and check access
	dataFileID := filepath.Base(req.URL.Path)
	df, err := s.File.ByID(dataFileID)
	switch {
	case err != nil:
		fmt.Printf("Failed looking up fileID %s: %s\n", dataFileID, err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case !s.Group.HasAccess(df.Owner, u.Email):
		fmt.Printf("No access owner: %s, accessed by: %s\n", df.Owner, u.Email)
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	default:
		var path string
		if isConvertedImage(df.MediaType.Mime) && download == "" {
			path = imageConversionPath(df.FileID())
		} else {
			path = mc.FilePath(df.FileID())
		}
		writer.Header().Set("Content-Type", df.MediaType.Mime)
		fmt.Printf("Serving path: %s\n", path)
		http.ServeFile(writer, req, path)
	}
}

// isTiff checks a name to see if it is for a TIFF file.
func isConvertedImage(mime string) bool {
	switch mime {
	case "image/tiff":
		return true
	case "image/x-ms-bmp":
		return true
	default:
		return false
	}
}

func imageConversionPath(id string) string {
	return filepath.Join(mc.FileDir(id), ".conversion", id+".jpg")
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

		m := util.NewGobMarshaler(conn)
		r := request.NewReqHandler(m, config.GetString("MCDIR"))
		go handleConnection(r, conn)
	}
}

// handleConnection handles connection requests by running the state machine. It also
// takes care of book keeping like shutting down the net and database connections when
// the connection is terminated.
func handleConnection(reqHandler *request.ReqHandler, conn net.Conn) {
	defer conn.Close()
	reqHandler.Run()
}
