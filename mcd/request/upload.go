package request

import (
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/protocol"
)

// upload handles the upload request. It validates the request and sends back the
// file ID to write to, and the offset to start sending from. The uploading of bytes
// is handled in the uploadLoop() method.
func (h *ReqHandler) upload(req *protocol.UploadReq) (*protocol.UploadResp, error) {
	dataFile, err := h.dai.File.ByID(req.DataFileID)
	if err != nil {
		return nil, mcerr.Errorm(mcerr.ErrNotFound, err)
	}

	if !h.dai.Group.HasAccess(dataFile.Owner, h.user) {
		return nil, mcerr.ErrNoAccess
	}

	dfLocationID := datafileLocationID(dataFile)
	fsize := datafileSize(h.mcdir, dfLocationID)

	switch {
	case fsize == -1:
		// Problem doing a stat on the file path, send back an error
		return nil, mcerr.Errorf(mcerr.ErrNoAccess, "Access to path for file %s denied", req.DataFileID)

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
		return nil, mcerr.Errorf(mcerr.ErrInvalid, "Invalid request: Expected size (%d) doesn't match the request size (%d).", dataFile.Size, req.Size)

	case dataFile.Checksum != req.Checksum:
		// Invalid request. The correct checksum was set at the time createFile was called.
		return nil, mcerr.Errorf(mcerr.ErrInvalid, "Invalid request: Expected checksum (%s) doesn't match the request checksum (%s).", dataFile.Checksum, req.Checksum)

	default:
		// We should never get here so this is a bug that we need to log
		return nil, mcerr.ErrInternal
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
		return 0, mcerr.Errorf(mcerr.ErrInvalid, "Fatal error fsize (%d) > ureqSize (%d) with equal checksums", fsize, reqSize)
	}
}
