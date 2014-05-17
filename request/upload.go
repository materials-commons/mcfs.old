package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/db"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/protocol"
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

type uploadHandler struct {
	w          io.WriteCloser
	dataFileID string
	nbytes     int64
	session    *r.Session
	*ReqHandler
}

func datafileWrite(w io.WriteCloser, bytes []byte) (int, error) {
	return w.Write(bytes)
}

func datafileClose(w io.WriteCloser, dataFileID string, session *r.Session) error {
	// Update datafile in db?
	w.Close()
	return nil
}

func datafileOpen(mcdir, dfid string, offset int64) (io.WriteCloser, error) {
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

/*
The following variables define functions for interacting with the datafile. They also
allow these functions to be replaced during testing when the test doesn't really
need to do anything with the datafile.
*/
var dfWrite = datafileWrite
var dfClose = datafileClose
var dfOpen = datafileOpen

func prepareUploadHandler(h *ReqHandler, dataFileID string, offset int64) (*uploadHandler, error) {
	f, err := dfOpen(h.mcdir, dataFileID, offset)
	if err != nil {
		return nil, err
	}

	session, _ := db.RSession()
	handler := &uploadHandler{
		w:          f,
		dataFileID: dataFileID,
		nbytes:     0,
		ReqHandler: h,
		session:    session,
	}

	return handler, nil
}

func (h *ReqHandler) uploadLoop(resp *protocol.UploadResp) reqStateFN {
	uploadHandler, err := prepareUploadHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		h.respError(nil, mc.Errorm(mc.ErrInternal, err))
		return h.nextCommand
	}

	h.respOk(resp)
	return uploadHandler.uploadState
}

func (h *uploadHandler) uploadState() reqStateFN {
	request := h.req()
	switch req := request.(type) {
	case protocol.SendReq:
		n, err := h.sendReqWrite(&req)
		if err != nil {
			dfClose(h.w, h.dataFileID, h.session)
			h.respError(nil, err)
			return h.nextCommand
		}
		h.nbytes = h.nbytes + int64(n)
		h.respOk(&protocol.SendResp{BytesWritten: n})
		return h.uploadState
	case errorReq:
		dfClose(h.w, h.dataFileID, h.session)
		return nil
	case protocol.LogoutReq:
		dfClose(h.w, h.dataFileID, h.session)
		h.respOk(&protocol.LogoutResp{})
		return h.startState
	case protocol.CloseReq:
		dfClose(h.w, h.dataFileID, h.session)
		return nil
	case protocol.DoneReq:
		dfClose(h.w, h.dataFileID, h.session)
		h.respOk(&protocol.DoneResp{})
		return h.nextCommand
	default:
		dfClose(h.w, h.dataFileID, h.session)
		return h.badRequestNext(mc.Errorf(mc.ErrInvalid, "Unknown Request Type %T", req))
	}
}

func (h *uploadHandler) sendReqWrite(req *protocol.SendReq) (int, error) {
	if req.DataFileID != h.dataFileID {
		return 0, mc.Errorf(mc.ErrInvalid, "Unexpected DataFileID %s, wanted: %s", req.DataFileID, h.dataFileID)
	}

	n, err := dfWrite(h.w, req.Bytes)
	if err != nil {
		return 0, mc.Errorf(mc.ErrInternal, "Write unexpectedly failed for %s", req.DataFileID)
	}

	return n, nil
}

func createDataFileDir(mcdir, dataFileID string) error {
	dirpath := datafileDir(mcdir, dataFileID)
	return os.MkdirAll(dirpath, 0777)
}
