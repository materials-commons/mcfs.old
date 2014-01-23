package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/contrib/schema"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/protocol"
	"io"
	"os"
)

type uploadReq struct {
	*protocol.UploadReq
	*ReqHandler
	dataFile *schema.DataFile
}

func (h *ReqHandler) upload(req *protocol.UploadReq) (*protocol.UploadResp, error) {
	ureq := &uploadReq{
		UploadReq:  req,
		ReqHandler: h,
	}

	resp := &protocol.UploadResp{}

	if err := ureq.isValid(); err != nil {
		return nil, err
	}

	fsize := datafileSize(h.mcdir, req.DataFileID)

	switch {
	case fsize == -1:
		// Problem doing a stat on the file path, send back an error
		return nil, fmt.Errorf("Access to path for file %s denied", req.DataFileID)
	case ureq.dataFile.Size == ureq.Size && ureq.dataFile.Checksum == ureq.Checksum:
		if fsize < ureq.Size {
			//interrupted transfer
			// send offset = fsize and ureq.dataFile.ID
			resp.DataFileID = req.DataFileID
			resp.Offset = fsize
		} else if fsize == ureq.Size {
			// nothing to send file upload completed
			resp.DataFileID = req.DataFileID
			resp.Offset = ureq.Size
		} else {
			// fsize > ureq.Size && checksums are equal
			// Houston we have a problem!
			return nil, fmt.Errorf("Fatal error fsize (%d) > ureq.Size (%d) with equal checksums", fsize, ureq.Size)
		}

	case ureq.dataFile.Size != ureq.Size:
		// wants to upload a new version
		if fsize < ureq.dataFile.Size {
			// Other upload hasn't completed - reject this one until other completes
			return nil, fmt.Errorf("Cannot create new version of data file when previous version hasn't completed loading.")
		} else {
			// create a new version and send new data file and offset = 0
			resp.DataFileID = ureq.createNewDataFileVersion()
			resp.Offset = 0
		}

	case ureq.dataFile.Size == ureq.Size && ureq.dataFile.Checksum != ureq.Checksum:
		// wants to upload new version
		if fsize < ureq.dataFile.Size {
			// Other upload hasn't completed - reject this one until other completes
			return nil, fmt.Errorf("Cannot create new version of data file when previous version hasn't completed loading.")
		} else {
			// create a new version start upload
			// send offset = 0 and a new datafile id
			resp.DataFileID = ureq.createNewDataFileVersion()
			resp.Offset = 0
		}

	default:
		// We should never get here so this is a bug that we need to log
		return nil, fmt.Errorf("Internal fatal error")
	}

	return resp, nil
}

func (req *uploadReq) isValid() error {
	dataFile, err := model.GetDataFile(req.DataFileID, req.session)
	switch {
	case err != nil:
		return fmt.Errorf("No such datafile %s", req.DataFileID)
	case !OwnerGaveAccessTo(dataFile.Owner, req.user, req.session):
		return fmt.Errorf("Permission denied to %s", req.DataFileID)
	default:
		req.dataFile = dataFile
		return nil
	}
}

func datafileSize(mcdir, dataFileID string) int64 {
	path := datafilePath(mcdir, dataFileID)
	finfo, err := os.Stat(path)
	switch {
	case err == nil:
		return finfo.Size()
	case os.IsNotExist(err):
		return 0
	default:
		return -1
	}
}

func (req *uploadReq) createNewDataFileVersion() (dataFileID string) {
	/*
		newDataFile := *req.dataFile
		newDataFile.Id = ""
		newDataFile.Parent = req.dataFile.Id
		rv, err := r.Table("datafiles").Insert(newDataFile).RunWrite(req.session)
		if err != nil {
			fmt.Println(err)
		}
		if rv.Inserted == 0 {
			fmt.Println("Nothing inserted!")
		}
		dataFileID = rv.GeneratedKeys[0]
		// Update datadir to point at new file
		var ddirs = []string{}
		for _, ddir := range req.dataFile.DataDirs {
			if ddir !=
		}
	*/
	return "NEW"
}

type uploadHandler struct {
	w          io.WriteCloser
	dataFileID string
	nbytes     int64
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

	handler := &uploadHandler{
		w:          f,
		dataFileID: dataFileID,
		nbytes:     0,
		ReqHandler: h,
	}

	return handler, nil
}

func (r *ReqHandler) uploadLoop(resp *protocol.UploadResp) ReqStateFN {
	if uploadHandler, err := prepareUploadHandler(r, resp.DataFileID, resp.Offset); err != nil {
		r.respError(err)
		return r.nextCommand
	} else {
		r.respOk(resp)
		return uploadHandler.uploadState
	}
}

func (h *uploadHandler) uploadState() ReqStateFN {
	request := h.req()
	switch req := request.(type) {
	case *protocol.SendReq:
		n, err := h.sendReqWrite(req)
		if err != nil {
			dfClose(h.w, h.dataFileID, h.session)
			h.respError(err)
			return h.nextCommand
		}
		h.nbytes = h.nbytes + int64(n)
		h.respOk(&protocol.SendResp{BytesWritten: n})
		return h.uploadState
	case *ErrorReq:
		dfClose(h.w, h.dataFileID, h.session)
		return nil
	case *protocol.LogoutReq:
		dfClose(h.w, h.dataFileID, h.session)
		h.respOk(&protocol.LogoutResp{})
		return h.startState
	case *protocol.CloseReq:
		dfClose(h.w, h.dataFileID, h.session)
		return nil
	case *protocol.DoneReq:
		dfClose(h.w, h.dataFileID, h.session)
		h.respOk(&protocol.DoneResp{})
		return h.nextCommand
	default:
		dfClose(h.w, h.dataFileID, h.session)
		return h.badRequestNext(fmt.Errorf("Unknown Request Type"))
	}
}

func (h *uploadHandler) sendReqWrite(req *protocol.SendReq) (int, error) {
	if req.DataFileID != h.dataFileID {
		return 0, fmt.Errorf("Unexpected DataFileID %s", req.DataFileID)
	}

	n, err := dfWrite(h.w, req.Bytes)
	if err != nil {
		return 0, fmt.Errorf("Write unexpectedly failed for %s", req.DataFileID)
	}

	return n, nil
}

func createDataFileDir(mcdir, dataFileID string) error {
	dirpath := datafileDir(mcdir, dataFileID)
	return os.MkdirAll(dirpath, 0777)
}
