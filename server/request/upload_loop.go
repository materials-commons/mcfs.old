package request

import (
	"crypto/md5"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/inuse"
	"github.com/materials-commons/mcfs/server/service"
	"io"
	"os"
)

// uploadFileHandler holds internal state and methods used by the upload loop.
type uploadFileHandler struct {
	w      io.WriteCloser
	file   *schema.File
	nbytes int64
	*ReqHandler
}

// uploadLoop sets up the loop to upload the files bytes.
func (h *ReqHandler) uploadLoop(resp *protocol.UploadResp) reqStateFN {
	uploadHandler, err := createUploadFileHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		inuse.Unmark(resp.DataFileID)
		h.respError(nil, mc.Errorm(mc.ErrInternal, err))
		return h.nextCommand
	}

	h.respOk(resp)
	return uploadHandler.uploadFile
}

// createUploadFileHandler creates an instance of the uploadHandler. This instance depends
// on having the file open. If it can't open the file it returns an error.
func createUploadFileHandler(h *ReqHandler, dataFileID string, offset int64) (*uploadFileHandler, error) {
	file, err := service.File.ByID(dataFileID)
	if err != nil {
		return nil, err
	}

	f, err := fileOpen(h.mcdir, file.FileID(), offset)
	if err != nil {
		return nil, err
	}

	handler := &uploadFileHandler{
		w:          f,
		file:       file,
		nbytes:     0,
		ReqHandler: h,
	}

	return handler, nil
}

// fileOpen opens the actual on disk file that the database file points to.
// It takes care of creating the directory structure if the file doesn't exist.
func fileOpen(mcdir, dfid string, offset int64) (io.WriteCloser, error) {
	path := datafilePath(mcdir, dfid)
	switch {
	case file.Exists(path):
		mode := os.O_RDWR
		if offset != 0 {
			mode = mode | os.O_APPEND
		}
		return os.OpenFile(path, mode, 0660)
	default:
		err := createDataFileDir(mcdir, dfid)
		if err != nil {
			return nil, err
		}
		return os.Create(path)
	}
}

// createDataFileDir creates the directory where a datafile is stored.
func createDataFileDir(mcdir, dataFileID string) error {
	dirpath := datafileDir(mcdir, dataFileID)
	return os.MkdirAll(dirpath, 0777)
}

// uploadFile performs the actual file upload. It accepts requests holding bytes
// and writes them to the file. At the moment this function is not optimized
// for speed. Each write requires a response back to the client before more bytes
// are sent.
func (u *uploadFileHandler) uploadFile() reqStateFN {
	request := u.req()
	switch req := request.(type) {
	case protocol.SendReq:
		return u.writeRequest(&req)
	case errorReq:
		u.fileClose()
		return nil
	case protocol.LogoutReq:
		u.fileClose()
		u.respOk(&protocol.LogoutResp{})
		return u.startState
	case protocol.CloseReq:
		u.fileClose()
		return nil
	case protocol.DoneReq:
		u.fileClose()
		u.respOk(&protocol.DoneResp{})
		return u.nextCommand
	default:
		u.fileClose()
		return u.badRequestNext(mc.Errorf(mc.ErrInvalid, "Unknown Request Type %T", req))
	}
}

// writeRequest writes the bytes to the file, checking status and validating
// that the write hasn't exceeded the expected size.
func (u *uploadFileHandler) writeRequest(req *protocol.SendReq) reqStateFN {
	n, err := u.sendReqWrite(req)
	u.nbytes = u.nbytes + int64(n)

	switch {
	case err != nil:
		// Problem writing to file.
		u.fileClose()
		u.respError(nil, err)
		return u.nextCommand

	case u.nbytes+u.file.Uploaded > u.file.Size:
		// Client is sending us more bytes than expecte file size.
		u.fileClose()
		u.respError(nil, mc.Errorf(mc.ErrInvalid, "Attempt to write more bytes to file than its expected size."))
		return u.nextCommand

	default:
		// No errors, continue accepting more bytes
		u.respOk(&protocol.SendResp{BytesWritten: n})
		return u.uploadFile
	}
}

// sendReqWrite writes bytes to the file.
func (u *uploadFileHandler) sendReqWrite(req *protocol.SendReq) (int, error) {
	if req.DataFileID != u.file.ID {
		return 0, mc.Errorf(mc.ErrInvalid, "Unexpected DataFileID %s, wanted: %s", req.DataFileID, u.file.ID)
	}

	n, err := u.fileWrite(req.Bytes)
	if err != nil {
		return 0, mc.Errorf(mc.ErrInternal, "Write unexpectedly failed for %s", req.DataFileID)
	}

	return n, nil
}

func (u *uploadFileHandler) fileWrite(bytes []byte) (int, error) {
	return u.w.Write(bytes)
}

type fileState int

const (
	// File completed upload, and the checksum verified.
	fileStateVerified fileState = iota

	// File upload was bad. This occurs when the checksums don't match
	// and the size of the uploaded file is equal to or greater than
	// the expected size.
	fileStateInvalid

	// File has not yet completed its upload.
	fileStateIncomplete
)

// fileClose closes the currently open file that bytes are being uploaded to. It
// also determines the state of the file. The state determines whether the file
// upload is complete, garbage and needs to be discarded, or is still a partial.
func (u *uploadFileHandler) fileClose() error {
	u.w.Close()
	switch status := u.fileState(); status {
	case fileStateVerified:
		// File has completed upload, and the checksum is correct.
		// Mark the file as current, as well as all files that point at it.
		u.markCurrent()
	case fileStateInvalid:
		// File has completed upload, but failed checksum verification. Return
		// an error and truncate the on disk version.
		u.truncate()
	default:
		// File hasn't completed uploading.
		u.updateUploaded()
	}
	inuse.Unmark(u.file.ID)
	return nil
}

// fileState determines an uploaded files state. It determines
// the state by comparing expected checksums and sizes.
func (u *uploadFileHandler) fileState() fileState {
	path := datafilePath(u.mcdir, u.file.FileID())
	checksum, err := file.HashStr(md5.New(), path)
	switch {
	case err != nil:
		return fileStateIncomplete
	case checksum == u.file.Checksum:
		return fileStateVerified
	default:
		// Not sure if the file is complete or not.
		// Look at the file size on disk vs expected
		// file size to determine file state.
		finfo, err := os.Stat(path)
		switch {
		case err != nil:
			return fileStateIncomplete
		case finfo.Size() > u.file.Size:
			// At this point we know the checksums don't match.
			// If the size is greater than or equal to what we
			// expect then the client sent us garbage.
			return fileStateInvalid
		default:
			// File size on disk < expected size, so not finished
			// uploading yet.
			return fileStateIncomplete
		}
	}
}

// markCurrent will mark the file being written to as current, plus
// all other files that point to it. It will hide all the files parents.
func (u *uploadFileHandler) markCurrent() {
	u.makeFileCurrent(u.file)
	files, _ := service.File.MatchOn("usesid", u.file.ID)
	for _, file := range files {
		// MatchOn query could return the current file if it has a usesid.
		// We don't want to update it twice because then it will add itself
		// to dependent objects twice.
		if file.ID != u.file.ID {
			u.makeFileCurrent(&file)
		}
	}
}

// makeFileCurrent goes through all the steps to take an entry that isn't connected
// into the system, make it the current entry, and hide its parent.
func (u *uploadFileHandler) makeFileCurrent(file *schema.File) {
	file.Uploaded = file.Size
	file.Current = true
	service.File.Update(file)
	service.File.AddDirectories(file, file.DataDirs...)
	if file.Parent != "" {
		f, err := service.File.ByID(file.Parent)
		if err == nil {
			service.File.Hide(f)
		}
	}
}

// truncate will make the size of the current file 0. This routine
// is used when an upload sends us garbage.
func (u *uploadFileHandler) truncate() {
	path := datafilePath(u.mcdir, u.file.ID)
	os.Truncate(path, 0)
}

// updateUploaded updates the total number of bytes written to the file.
func (u *uploadFileHandler) updateUploaded() {
	u.file.Uploaded += u.nbytes
	service.File.Update(u.file)
}
