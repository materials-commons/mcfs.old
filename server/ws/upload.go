package ws

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

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

// A FlowRequest encapsulates the flowjs protocol for uploading a file. The
// protocol supports extensions to the protocol. We extend the protocol to
// include Materials Commons specific information. It is also expected that
// the data sent by flow or another client will be placed in chunkData.
type FlowRequest struct {
	FlowChunkNumber  int32           `json:"flowChunkNumber"`  // The chunk being sent.
	FlowTotalChunks  int32           `json:"flowTotalChunks"`  // The total number of chunks to send.
	FlowChunkSize    int32           `json:"flowChunkSize"`    // The size of the chunk.
	FlowTotalSize    int64           `json:"flowTotalSize"`    // The size of the file being uploaded.
	FlowIdentifier   string          `json:"flowIdentifier"`   // A unique identifier used by Flow. Not guaranteed to be a GUID.
	FlowFileName     string          `json:"flowFilename"`     // The file name being uploaded.
	FlowRelativePath string          `json:"flowRelativePath"` // When available the relative file path.
	ProjectID        string          `json:"projectID"`        // Materials Commons Project ID.
	DirectoryID      string          `json:"directoryID"`      // Materials Commons Directory ID.
	FileID           string          `json:"fileID"`           // Materials Commons File ID.
	Chunk            *multipart.Part `json:"-"`                // The file data.
}

func newUploadResource(container *restful.Container) error {
	uploadResource := uploadResource{assembleRequest: make(chan finishRequest, 50)}
	uploadResource.startAssemblers()
	uploadResource.register(container)
	return nil
}

func (r uploadResource) register(container *restful.Container) {
	ws := new(restful.WebService)

	ws.Path("/upload").
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/file").To(r.uploadFileChunk).
		Consumes("multipart/form-data").
		Doc("Upload file"))
	ws.Route(ws.GET("/file").To(r.restartFileChunk).
		Doc("Restart upload").
		Reads(FlowRequest{}))

	container.Add(ws)
}

func (r uploadResource) restartFileChunk(request *restful.Request, response *restful.Response) {
	response.WriteErrorString(http.StatusInternalServerError, "no such file")
	//response.WriteErrorString(http.StatusOK, "")
}

func (r uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response) {
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, fmt.Sprintf("Bad Request: %s", err))
		return
	}

	uploadPath := fileUploadPath(flowRequest.ProjectID, flowRequest.DirectoryID, flowRequest.FileID)
	if err := os.MkdirAll(uploadPath, chunkPerms); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, fmt.Sprintf("Unable to create temporary chunk space: %s", err))
		return
	}

	// cpath := chunkPath(uploadPath, flowRequest.FlowIdentifier)
	// if err := ioutil.WriteFile(cpath, flowRequest.ChunkData, chunkPerms); err != nil {
	// 	response.WriteErrorString(http.StatusInternalServerError, fmt.Sprintf("Unable to write chunk for file: %s", err))
	// 	return
	// }

	if flowRequest.FlowChunkNumber == flowRequest.FlowTotalChunks {
		r.finishUpload(uploadPath)
	}

	response.WriteErrorString(http.StatusOK, "")
}

func (r uploadResource) finishUpload(uploadPath string) {

}

func fileUploadPath(projectID, directoryID, fileID string) string {
	return filepath.Join(mc.Dir(), "upload", projectID, directoryID, fileID)
}

func chunkPath(uploadPath, chunkID string) string {
	return filepath.Join(uploadPath, chunkID)
}
