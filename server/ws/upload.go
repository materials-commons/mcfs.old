package ws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcfs/base/log"
)

const chunkPerms = 0700 // Permissions to set uploads to

// An uploadResource handles all upload requests.
type uploadResource struct {
	ctracker *chunkTracker
	log      log15.Logger // Resource specific logging.
}

// A FlowRequest encapsulates the flowjs protocol for uploading a file. The
// protocol supports extensions to the protocol. We extend the protocol to
// include Materials Commons specific information. It is also expected that
// the data sent by flow or another client will be placed in chunkData.
type FlowRequest struct {
	FlowChunkNumber  int32  `json:"flowChunkNumber"`  // The chunk being sent.
	FlowTotalChunks  int32  `json:"flowTotalChunks"`  // The total number of chunks to send.
	FlowChunkSize    int32  `json:"flowChunkSize"`    // The size of the chunk.
	FlowTotalSize    int64  `json:"flowTotalSize"`    // The size of the file being uploaded.
	FlowIdentifier   string `json:"flowIdentifier"`   // A unique identifier used by Flow. Not guaranteed to be a GUID.
	FlowFileName     string `json:"flowFilename"`     // The file name being uploaded.
	FlowRelativePath string `json:"flowRelativePath"` // When available the relative file path.
	ProjectID        string `json:"projectID"`        // Materials Commons Project ID.
	DirectoryID      string `json:"directoryID"`      // Materials Commons Directory ID.
	FileID           string `json:"fileID"`           // Materials Commons File ID.
	Chunk            []byte `json:"-"`                // The file data.
}

// newUploadResource creates a new instance of the upload resource, including
// registering its paths.
func newUploadResource(container *restful.Container) error {
	uploadResource := uploadResource{
		ctracker: newChunkTracker(),
		log:      log.New("resource", "upload"),
	}
	uploadResource.register(container)
	return nil
}

// register registers the resource and its paths.
func (r uploadResource) register(container *restful.Container) {
	ws := new(restful.WebService)

	ws.Path("/upload").
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/chunk").To(r.uploadFileChunk).
		Consumes("multipart/form-data").
		Doc("Upload a file chunk"))
	ws.Route(ws.GET("/chunk").To(r.testFileChunk).
		Reads(FlowRequest{}).
		Doc("Test if chunk already uploaded."))

	container.Add(ws)
}

// testFileChunk checks if a chunk has already been uploaded. At the moment we don't
// support this functionality, so always return an error.
func (r uploadResource) testFileChunk(request *restful.Request, response *restful.Response) {
	response.WriteErrorString(http.StatusInternalServerError, "no such file")
}

// uploadFileChunk uploads a new file chunk.
func (r uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response) {
	// Create request
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		r.log.Error(log.Msg("Error converting form to FlowRequest: %s", err))
		response.WriteErrorString(http.StatusNotAcceptable, fmt.Sprintf("Bad Request: %s", err))
		return
	}

	// Ensure directory path exists
	uploadPath, err := r.createUploadDir(flowRequest)
	if err != nil {
		msg := fmt.Sprintf("Unable to create temporary chunk space: %s", err)
		r.log.Error(msg)
		response.WriteErrorString(http.StatusInternalServerError, msg)
		return
	}

	// Write chunk and determine if done.
	if err := r.processChunk(uploadPath, flowRequest); err != nil {
		msg := fmt.Sprintf("Unable to write chunk for file: %s", err)
		r.log.Error(msg)
		response.WriteErrorString(http.StatusInternalServerError, msg)
		return
	}

	response.WriteErrorString(http.StatusOK, "")
}

// createUploadDir creates the directory for the chunk.
func (r uploadResource) createUploadDir(flowRequest *FlowRequest) (string, error) {
	uploadPath := fileUploadPath(flowRequest.ProjectID, flowRequest.DirectoryID, flowRequest.FileID)

	// os.MkdirAll returns nil if the path already exists.
	return uploadPath, os.MkdirAll(uploadPath, chunkPerms)
}

// processChunk writes the chunk and determines if this is the last chunk to write.
// If the last chunk has been uploaded it kicks off a reassembly of the file.
func (r uploadResource) processChunk(uploadPath string, flowRequest *FlowRequest) error {
	cpath := chunkPath(uploadPath, flowRequest.FlowChunkNumber)
	if err := r.writeChunk(cpath, flowRequest.Chunk); err != nil {
		return err
	}

	if r.uploadDone(flowRequest) {
		r.finishUpload(flowRequest)
	}

	return nil
}

// uploadDone checks to see if the upload has finished.
func (r uploadResource) uploadDone(flowRequest *FlowRequest) bool {
	id := r.uploadID(flowRequest)
	count := r.ctracker.addChunk(id)
	return count == flowRequest.FlowTotalChunks
}

// uploadID creates the unique id for this files upload request
func (r uploadResource) uploadID(flowRequest *FlowRequest) string {
	return fmt.Sprintf("%s-%s-%s", flowRequest.ProjectID, flowRequest.DirectoryID, flowRequest.FileID)
}

// writeChunk writes a file chunk.
func (r uploadResource) writeChunk(chunkpath string, chunk []byte) error {
	return ioutil.WriteFile(chunkpath, chunk, chunkPerms)
}

// finishUpload marks the upload as finished and kicks off an assembler to assemble the file.
func (r uploadResource) finishUpload(flowRequest *FlowRequest) {
	id := r.uploadID(flowRequest)
	r.ctracker.clear(id)
	assembler := newAssemberFromFlowRequest(flowRequest)
	go assembler.assembleFile()
}
