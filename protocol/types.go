package protocol

import "time"

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
	FileID string // The internal id of the file.
}

// SyncStartResp contains the token to use for syncing a project.
type SyncStartResp struct {
	TokenID string // The token to use on all sync requests
}

// SyncStatusResp contains the current sync status for a project.
type SyncStatusResp struct {
	ProjectID string    // The project ID being sync
	Started   time.Time // The time the sync started
	User      string    // The user performing the sync
}
