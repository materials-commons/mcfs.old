package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request/inprogress"
	"github.com/materials-commons/mcfs/service"
	"io"
	"os"
)

// upload handles the upload request. It validates the request and sends back the
// file ID to write to, and the offset to start sending from. The uploading of bytes
// is handled in the uploadLoop() method.
func (h *ReqHandler) upload(req *protocol.UploadReq) (*protocol.UploadResp, error) {
	dataFile, err := service.File.ByID(req.DataFileID)
	if err != nil {
		return nil, mc.Errorm(mc.ErrNotFound, err)
	}

	if !service.Group.HasAccess(dataFile.Owner, h.user) {
		return nil, mc.ErrNoAccess
	}

	dfLocationID := datafileLocationID(dataFile)
	fsize := datafileSize(h.mcdir, dfLocationID)

	switch {
	case fsize == -1:
		// Problem doing a stat on the file path, send back an error
		return nil, mc.Errorf(mc.ErrNoAccess, "Access to path for file %s denied", req.DataFileID)

	case dataFile.Size == req.Size && dataFile.Checksum == req.Checksum:
		// Request looks ok, determine offset to use.
		var offset int64
		if offset, err = responseOffset(fsize, req.Size); err != nil {
			return nil, err
		}
		dfid := dfLocationID

		// If there is nothing to write then we send back the original id.
		// Otherwise, if we have bytes to write we send back the id to
		// write to. Since the file could point to another file that is
		// a duplicate (but hasn't been completely uploaded), the id could
		// be different as it could point to the duplicate.
		if offset == dataFile.Size {
			dfid = dataFile.ID
		}

		if inprogress.Mark(dfid) {
			// Attempt to mark file as in progress. If Mark returns true then
			// the file was already in progress, so return error.
			return nil, mc.Errorf(mc.ErrInvalid, "File upload already in progress: %s", dataFile.ID)
		}
		return &protocol.UploadResp{DataFileID: dfid, Offset: offset}, nil

	case dataFile.Size != req.Size:
		// Invalid request. The correct size was set at the time createFile was called.
		return nil, mc.Errorf(mc.ErrInvalid, "Invalid request: Expected size (%d) doesn't match the request size (%d).", dataFile.Size, req.Size)

	case dataFile.Checksum != req.Checksum:
		// Invalid request. The correct checksum was set at the time createFile was called.
		return nil, mc.Errorf(mc.ErrInvalid, "Invalid request: Expected checksum (%s) doesn't match the request checksum (%s).", dataFile.Checksum, req.Checksum)

	default:
		// We should never get here so this is a bug that we need to log
		return nil, mc.ErrInternal
	}
}

// responseOffset determines the offset to start sending bytes from. This
// call assumes that the checksums have been validated.
func responseOffset(fsize, reqSize int64) (int64, error) {
	switch {
	case fsize < reqSize:
		// interrupted transfer, send offset = fsize. Thus the
		// client will start sending from fsize, thereby sending
		// the rest of the bytes for the file.
		return fsize, nil
	case fsize == reqSize:
		// No bytes need to be sent. Tell client the number of bytes
		// to upload is exactly equal to the file. This will cause
		// the client to skip sending anything.
		return reqSize, nil
	default:
		// fsize > reqSize. This is a problem on the client side.
		return 0, mc.Errorf(mc.ErrInvalid, "Fatal error fsize (%d) > ureqSize (%d) with equal checksums", fsize, reqSize)
	}
}

/* ********************* upload loop section ********************* */

// uploadFileHandler holds internal state and methods used by the upload loop.
type uploadFileHandler struct {
	w          io.WriteCloser
	dataFileID string
	nbytes     int64
	*ReqHandler
}

// uploadLoop sets up the loop to upload the files bytes.
func (h *ReqHandler) uploadLoop(resp *protocol.UploadResp) reqStateFN {
	uploadHandler, err := createUploadFileHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		inprogress.Unmark(resp.DataFileID)
		h.respError(nil, mc.Errorm(mc.ErrInternal, err))
		return h.nextCommand
	}

	h.respOk(resp)
	return uploadHandler.uploadFile
}

// createUploadFileHandler creates an instance of the uploadHandler. This instance depends
// on having the file open. If it can't open the file it returns an error.
func createUploadFileHandler(h *ReqHandler, dataFileID string, offset int64) (*uploadFileHandler, error) {
	f, err := fileOpen(h.mcdir, dataFileID, offset)
	if err != nil {
		return nil, err
	}

	handler := &uploadFileHandler{
		w:          f,
		dataFileID: dataFileID,
		nbytes:     0,
		ReqHandler: h,
	}

	return handler, nil
}

// uploadFile performs the actual file upload. It accepts requests holding bytes
// and writes them to the file. At the moment this function is not optimized
// for speed. Each write requires a response back to the client before more bytes
// are sent.
func (u *uploadFileHandler) uploadFile() reqStateFN {
	request := u.req()
	switch req := request.(type) {
	case protocol.SendReq:
		n, err := u.sendReqWrite(&req)
		if err != nil {
			u.fileClose()
			u.respError(nil, err)
			return u.nextCommand
		}
		u.nbytes = u.nbytes + int64(n)
		u.respOk(&protocol.SendResp{BytesWritten: n})
		return u.uploadFile
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

func (u *uploadFileHandler) fileWrite(bytes []byte) (int, error) {
	return u.w.Write(bytes)
}

func (u *uploadFileHandler) fileClose() error {
	// Update datafile in db?
	u.w.Close()
	inprogress.Unmark(u.dataFileID)
	return nil
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

// sendReqWrite writes bytes to the file.
func (u *uploadFileHandler) sendReqWrite(req *protocol.SendReq) (int, error) {
	if req.DataFileID != u.dataFileID {
		return 0, mc.Errorf(mc.ErrInvalid, "Unexpected DataFileID %s, wanted: %s", req.DataFileID, u.dataFileID)
	}

	n, err := u.fileWrite(req.Bytes)
	if err != nil {
		return 0, mc.Errorf(mc.ErrInternal, "Write unexpectedly failed for %s", req.DataFileID)
	}

	return n, nil
}

// createDataFileDir creates the directory where a datafile is stored.
func createDataFileDir(mcdir, dataFileID string) error {
	dirpath := datafileDir(mcdir, dataFileID)
	return os.MkdirAll(dirpath, 0777)
}
