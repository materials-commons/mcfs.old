package flow

import "fmt"

// A FlowRequest encapsulates the flowjs protocol for uploading a file. The
// protocol supports extensions to the protocol. We extend the protocol to
// include Materials Commons specific information. It is also expected that
// the data sent by flow or another client will be placed in chunkData.
type Request struct {
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
	ChunkHash        string `json:"chunkHash"`        // The computed MD5 hash for the chunk (optional).
	FileHash         string `json:"fileHash"`         // The computed MD5 hash for the file (optional)
}

func (r *Request) UploadID() string {
	return fmt.Sprintf("%s-%s-%s", r.ProjectID, r.DirectoryID, r.FileID)
}
