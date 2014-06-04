package protocol

import (
	"github.com/materials-commons/mcfs/base/mcerr"
	"time"
)

const (
	// LoginRequest LoginReq
	LoginRequest uint8 = iota

	// LoginResponse LoginResp
	LoginResponse

	// LogoutRequest LogoutReq
	LogoutRequest

	// CreateProjectRequest CreateProjectReq
	CreateProjectRequest

	// CreateProjectResponse CreateProjectResp
	CreateProjectResponse

	// CreateDirectoryRequest CreateDirectoryReq
	CreateDirectoryRequest

	// CreateDirectoryResponse CreateDirectoryResp
	CreateDirectoryResponse

	// CreateFileRequest CreateFileReq
	CreateFileRequest

	// CreateFileResponse CreateFileResp
	CreateFileResponse

	// DirectoryStatRequest DirectoryStatReq
	DirectoryStatRequest

	// DirectoryStatResponse DirectoryStatResp
	DirectoryStatResponse

	// SendBytesRequest SendBytesReq
	SendBytesRequest
)

// Status is the status of the request. All response type messages include a request status.
// This denotes success or failure of the request. On failure a message with additional details
// may be included.
type Status struct {
	Status  mcerr.ErrorCode // Error code (can be translated to an error)
	Message string       // Additional status message
}

// LoginReq is sent to login to the server and prepare to issue commands.
type LoginReq struct {
	User   string // User login, usually their email address
	APIKey string // The APIKey the server holds
}

// LoginResp will inform the client whether the login request was successful.
type LoginResp struct {
	Status
}

// LogoutReq is sent to end a session and disconnect. There client doesn't wait for
// a response from the server. It terminates the connection after sending this request.
type LogoutReq struct {
}

// CreateProjectReq is sent to create a project.
// If shared is set to true then the server will check if a project
// matching this name exists in any of the projects user has access to. If so
// it will use that project.
type CreateProjectReq struct {
	Name   string // The name of the project to create
	Shared bool   // Should we check shared projects?
}

// CreateProjectResp is the response to a CreateProjectRequest. A request to create
// a project will not create a project if a matching project already exists. In that case
// it will indicate this in the Status field.
//
// The status of the request. There are two error codes for success:
//    ErrorCodeSuccess - Project was created
//    ErrorCodeExists  - Project already exists
// All other error codes are failures.
type CreateProjectResp struct {
	Status      Status // Request status
	ProjectID   string // The internal ProjectID for the created or existing project.
	DirectoryID string // The internal id for the directory that the project is stored in.
}

// CreateDirectoryReq is sent to create a directory in a project.
type CreateDirectoryReq struct {
	ProjectID string // The project to create the directory in.
	Path      string // The directory path, relative to the project to create. All members of the path except the leaf must exist.
}

// CreateDirectoryResp is the response for a CreateDirectoryRequest. A request to create
// a directory will not create a directory if a matching directory already exists. In
// that case it will indicate this in the Status field.
//
// The status of the request. There are two error codes for success:
//    ErrorCodeSuccess - Directory was created
//    ErrorCodeExists  - Directory already exists
// All other error codes are failures.
type CreateDirectoryResp struct {
	Status             // Request status
	DirectoryID string // The internal id of the directory.
}

// CreateFileReq is sent to create a new file on the server. It is expected that
// after this request is sent that an attempt will be made to upload the file. Until
// this upload succeeds newer versions of the file cannot be created.
//
// If CreateNewVersion is set to true then a new file version will
// be created if the previous file has already been successfully
// uploaded.
type CreateFileReq struct {
	ProjectID        string // The id of the project to create the file in.
	DirectoryID      string // The id of the directory in the project to create the file in.
	Name             string // The name of the file.
	Checksum         string // The files MD5 hash
	Size             int64  // The size of the file
	CreateNewVersion bool   // Should we create a new version
}

// CreateFileResp is the response to a CreateFileRequest.
//
// Status of the request. There are three error codes for success:
//    ErrorCodeSuccess - File was created
//    ErrorCodeExists  - File already exists
//    ErrorCodeNew     - A new version of the file was created
// All other error codes are failures.
type CreateFileResp struct {
	Status        // Request status
	FileID string // The internal id of the file.
}

// DirectoryStatReq requests the server to send back its current
// view of the given directory.
type DirectoryStatReq struct {
	ProjectID   string // Project ID on server that the directory is in
	DirectoryID string // Directory ID on server
}

// StatEntryType identifies the type of stat entry as either a file
// or a directory.
type StatEntryType uint8

const (
	// StatTypeDirectory entry is a directory
	StatTypeDirectory StatEntryType = iota

	// StatTypeFile entry is a file
	StatTypeFile
)

// StatEntry describes a single file or directory.
//
// If Type is StatTypeDirectory, then Checksum, Size, and UploadedSize are not defined.
type StatEntry struct {
	Type         StatEntryType // The type of entry
	Name         string        // Name of entry
	ID           string        // The internal ID of the entry
	Owner        string        // Owner of entry
	Checksum     string        // The computed MD5 Hash
	Size         int64         // The size of the file
	UploadedSize int64         // The actual uploaded size of the file
	Birthtime    time.Time     // Datetime the entry was created on the server.
}

// DirectoryStatResp is the response for a DirectoryStatRequest. It returns
// a list of all the known entries in the directory.
type DirectoryStatResp struct {
	Status                  // Status of the request
	ProjectID   string      // ProjectID passed in the request
	DirectoryID string      // DirectoryID passed in the request
	Entries     []StatEntry // A list of all the entries for this directory.
}

// SendBytesReq contains the bytes to write.
type SendBytesReq struct {
	FileID string // File ID on the server that these bytes are for
	Bytes  []byte // The set of bytes to write
}
