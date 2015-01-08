package web

import (
	"fmt"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcfs/base/log"
	"github.com/materials-commons/mcfs/mcd/dai"
	"github.com/materials-commons/mcfs/mcd/request"
	"github.com/materials-commons/mcfs/mcd/servers"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type webServer struct{}

var l = log.New("server", "WebServer")
var server = &webServer{}

var mcdir = config.GetString("MCDIR")

func Server() servers.Instance {
	return server
}

func (s *webServer) Init() {

}

func (s *webServer) Run(stopChan <-chan struct{}) {
	go s.webserver()
	<-stopChan
}

func (s *webServer) webserver() {
	port := config.GetInt("MCFS_HTTP_PORT")
	http.HandleFunc("/datafiles/static/", s.datafileHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	var _ = err
}

func (s *webServer) datafileHandler(writer http.ResponseWriter, req *http.Request) {
	apikey := req.FormValue("apikey")
	if apikey == "" {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	download := req.FormValue("download")

	// Verify key
	u, err := dai.User.ByAPIKey(apikey)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get datafile from db and check access
	dataFileID := filepath.Base(req.URL.Path)
	df, err := dai.File.ByID(dataFileID)
	switch {
	case err != nil:
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case !dai.Group.HasAccess(df.Owner, u.Email):
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	default:
		var path string
		if isTiff(df.Name) && download == "" {
			path = tiffConversionPath(mcdir, df.FileID())
		} else {
			extension := strings.ToLower(filepath.Ext(df.Name))
			if mimetype := mime.TypeByExtension(extension); mimetype != "" {
				writer.Header().Set("Content-Type", mimetype)
			}
			path = request.DataFilePath(mcdir, df.FileID())
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
