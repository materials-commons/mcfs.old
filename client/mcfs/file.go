package mcfs

import (
	"crypto/md5"
	"fmt"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/protocol"
	"io"
	"os"
	"path/filepath"
)

// RestartFileUpload restarts a partially completed upload.
func (c *Client) RestartFileUpload(dataFileID, path string) (bytesUploaded int64, err error) {
	checksum, size, err := fileInfo(path)
	if err != nil {
		return 0, err
	}

	return c.uploadFile(dataFileID, path, checksum, size)
}

// UploadNewFile uploads a new file to the server.
func (c *Client) UploadNewFile(projectID, dataDirID, path string) (bytesUploaded int64, dataFileID string, err error) {
	checksum, size, err := fileInfo(path)
	if err != nil {
		return 0, "", err
	}

	createFileReq := &protocol.CreateFileReq{
		ProjectID: projectID,
		DataDirID: dataDirID,
		Name:      file.NormalizePath(filepath.Base(path)),
		Checksum:  checksum,
		Size:      size,
	}

	dataFileID, err = c.createFile(createFileReq)
	if err != nil {
		return 0, "", err
	}

	n, err := c.uploadFile(dataFileID, path, checksum, size)
	return n, dataFileID, err
}

func fileInfo(path string) (checksum string, size int64, err error) {
	checksum, err = file.HashStr(md5.New(), path)
	if err != nil {
		return
	}

	finfo, err := os.Stat(path)
	if err != nil {
		return
	}
	size = finfo.Size()
	return
}

func (c *Client) createFile(req *protocol.CreateFileReq) (dataFileID string, err error) {
	resp, err := c.doRequest(*req)
	if err != nil {
		return "", err
	}

	switch t := resp.(type) {
	case protocol.CreateResp:
		return t.ID, nil
	default:
		fmt.Printf("3 %s %T\n", ErrBadResponseType, t)
		return "", ErrBadResponseType
	}
}

func (c *Client) uploadFile(dataFileID, path, checksum string, size int64) (bytesUploaded int64, err error) {
	uploadReq := &protocol.UploadReq{
		DataFileID: dataFileID,
		Checksum:   checksum,
		Size:       size,
	}

	uploadResp, err := c.startUpload(uploadReq)
	switch {
	case err != nil:
		return 0, err
	//case uploadResp.DataFileID != dataFileID:
	//	return 0, fmt.Errorf("DataFileIDs don't match %d %#v %s", size, uploadResp, dataFileID)
	default:
		if uploadResp.DataFileID != dataFileID {
			fmt.Printf("Using an existing datafile %s for id %s\n", uploadResp.DataFileID, dataFileID)
		}
		n, err := c.sendFile(uploadResp.DataFileID, path, uploadResp.Offset)
		c.endUpload()
		return n, err
	}

}

func (c *Client) startUpload(req *protocol.UploadReq) (*protocol.UploadResp, error) {
	resp, err := c.doRequest(*req)
	if err != nil {
		return nil, err
	}

	switch t := resp.(type) {
	case protocol.UploadResp:
		return &t, nil
	default:
		fmt.Printf("4 %s %T\n", ErrBadResponseType, t)
		return nil, ErrBadResponseType
	}
}

func (c *Client) endUpload() {
	c.doRequest(&protocol.DoneReq{})
}

func (c *Client) sendFile(dataFileID, path string, offset int64) (bytesSent int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	_, err = f.Seek(offset, 0)
	if err != nil {
		return 0, err
	}

	return c.sendFileBytes(f, dataFileID)
}

func (c *Client) sendFileBytes(f *os.File, dataFileID string) (totalSent int64, err error) {
	sendReq := protocol.SendReq{
		DataFileID: dataFileID,
	}

	buf := make([]byte, readBufSize)
	for {
		var bytesSent int
		n, err := f.Read(buf)
		if n != 0 {
			sendReq.Bytes = buf[:n]
			bytesSent, err = c.sendBytes(&sendReq)
			if err != nil {
				break
			}
			totalSent = totalSent + int64(bytesSent)
		}
		if err != nil {
			break
		}
	}

	if err != nil && err != io.EOF {
		return totalSent, err
	}

	return totalSent, nil
}

func (c *Client) sendBytes(sendReq *protocol.SendReq) (bytesSent int, err error) {
	resp, err := c.doRequest(sendReq)
	if err != nil {
		return 0, err
	}

	switch t := resp.(type) {
	case protocol.SendResp:
		return t.BytesWritten, nil
	default:
		fmt.Printf("5 %s %T\n", ErrBadResponseType, t)
		return 0, ErrBadResponseType
	}
}
