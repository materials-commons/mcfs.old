package ws

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/base/mc"
)

const chunkPerms = 0700
const maxAssemblers = 32

type finishRequest struct {
	projectID   string
	directoryID string
	fileID      string
	uploadPath  string
}

type uploadResource struct {
	assembleRequest chan finishRequest
}

type FlowRequest struct {
	FlowChunkNumber  int64  `json:"flowChunkNumber"`
	FlowTotalChunks  int64  `json:"flowTotalChunks"`
	FlowChunkSize    int32  `json:"flowChunkSize"`
	FlowTotalSize    int64  `json:"flowTotalSize"`
	FlowIdentifier   string `json: "flowIdentifier"`
	FlowFileName     string `json: "flowFilename"`
	FlowRelativePath string `json: "flowRelativePath"`
	ProjectID        string `json:"projectID"`
	DirectoryID      string `json:"directoryID"`
	FileID           string `json: "fileID"`
	ChunkData        []byte `json: "chunkData" `
}

func newUploadResource(container *restful.Container) error {
	uploadResource := uploadResource{assembleRequest: make(chan finishRequest, 50)}
	uploadResource.startAssemblers()
	uploadResource.register(container)
	return nil
}

func (r uploadResource) startAssemblers() {
	for i := 0; i < maxAssemblers; i++ {
		go r.fileAssembler()
	}
}

func (r uploadResource) fileAssembler() {
	for request := range r.assembleRequest {
		r.assembleFile(request)
	}
}

type byChunk []os.FileInfo

func (c byChunk) Len() int      { return len(c) }
func (c byChunk) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c byChunk) Less(i, j int) bool {
	chunkIName, _ := strconv.Atoi(c[i].Name())
	chunkJName, _ := strconv.Atoi(c[j].Name())
	return chunkIName < chunkJName
}

func (r uploadResource) assembleFile(request finishRequest) {

	// reassemble file
	filePath := mc.FilePath(request.fileID)
	fdst, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer fdst.Close()

	finfos, err := ioutil.ReadDir(request.uploadPath)
	if err != nil {
		return
	}

	sort.Sort(byChunk(finfos))
	for _, finfo := range finfos {
		fsrc, err := os.Open(chunkPath(request.uploadPath, finfo.Name()))
		if err != nil {
			return
		}
		defer fsrc.Close()
		io.Copy(fdst, fsrc)
	}
	os.RemoveAll(request.uploadPath)
}

func (r uploadResource) register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/upload").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("file").To(r.uploadFileChunk).
		Doc("Upload file").
		Reads(FlowRequest{}))
}

func (r uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response) {
	flowRequest := new(FlowRequest)

	if err := request.ReadEntity(&flowRequest); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, fmt.Sprintf("Bad Request: %s", err))
		return
	}

	uploadPath := fileUploadPath(flowRequest.ProjectID, flowRequest.DirectoryID, flowRequest.FileID)
	if err := os.MkdirAll(uploadPath, chunkPerms); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, fmt.Sprintf("Unable to create temporary chunk space: %s", err))
		return
	}

	cpath := chunkPath(uploadPath, flowRequest.FlowIdentifier)
	if err := ioutil.WriteFile(cpath, flowRequest.ChunkData, chunkPerms); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, fmt.Sprintf("Unable to write chunk for file: %s", err))
		return
	}

	if flowRequest.FlowChunkNumber == flowRequest.FlowTotalChunks {
		r.finishUpload(uploadPath)
	}

}

func (r uploadResource) finishUpload(uploadPath string) {

}

func fileUploadPath(projectID, directoryID, fileID string) string {
	return filepath.Join(mc.Dir(), "upload", projectID, directoryID, fileID)
}

func chunkPath(uploadPath, chunkID string) string {
	return filepath.Join(uploadPath, chunkID)
}
