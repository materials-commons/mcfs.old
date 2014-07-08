package ws

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strconv"

	"github.com/emicklei/go-restful"
)

var (
	ErrNoParam  = errors.New("no param")
	ErrBadParam = errors.New("bad param")
)

func getParam(name string, request *restful.Request) (param string) {
	param = request.PathParameter(name)
	if param != "" {
		return param
	}

	attr := request.Attribute(name)
	if attr != nil {
		param, _ = attr.(string)
	}
	return param
}

func getParamInt32(name string, request *restful.Request) (int32, error) {
	param := getParam(name, request)
	switch {
	case param == "":
		return 0, ErrNoParam
	default:
		i, err := strconv.ParseInt(param, 0, 32)
		if err != nil {
			return 0, ErrBadParam
		}
		return int32(i), nil
	}
}

func getParamInt64(name string, request *restful.Request) (int64, error) {
	param := getParam(name, request)
	switch {
	case param == "":
		return 0, ErrNoParam
	default:
		i, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return 0, ErrBadParam
		}
		return int64(i), nil
	}
}

func getFlowRequest(request *restful.Request) (flowRequest FlowRequest, err error) {

	return
}

func form2FlowRequest(request *restful.Request) (*FlowRequest, error) {
	var (
		r      FlowRequest
		err    error
		reader *multipart.Reader
		part   *multipart.Part
	)
	buf := new(bytes.Buffer)
	reader, err = request.Request.MultipartReader()
	if err != nil {
		return nil, err
	}

	for {
		part, err = reader.NextPart()
		if err != nil {
			break
		}

		name := part.FormName()
		if name != "chunkData" {
			io.Copy(buf, part)
		}
		switch name {
		case "flowChunkNumber":
			r.FlowChunkNumber = atoi32(buf.String())
		case "flowTotalChunks":
			r.FlowTotalChunks = atoi32(buf.String())
		case "flowChunkSize":
			r.FlowChunkSize = atoi32(buf.String())
		case "flowTotalSize":
			r.FlowTotalSize = atoi64(buf.String())
		case "flowIdentifier":
			r.FlowIdentifier = buf.String()
		case "flowFileName":
			r.FlowFileName = buf.String()
		case "flowRelativePath":
			r.FlowRelativePath = buf.String()
		case "projectID":
			r.ProjectID = buf.String()
		case "directoryID":
			r.DirectoryID = buf.String()
		case "fileID":
			r.FileID = buf.String()
		case "chunkData":
			if r.Chunk, err = ioutil.ReadAll(part); err != nil {
				fmt.Print("ReadAll error =", err)
			}
		}
		buf.Reset()
	}

	if err != io.EOF {
		return nil, err
	}

	return &r, nil
}

func atoi64(str string) int64 {
	i, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return -1
	}

	return i
}

func atoi32(str string) int32 {
	i := atoi64(str)
	return int32(i)
}

func flowRequest2FinishRequest(flowRequest *FlowRequest) finishRequest {
	return finishRequest{
		projectID:   flowRequest.ProjectID,
		directoryID: flowRequest.DirectoryID,
		fileID:      flowRequest.FileID,
		uploadPath:  fileUploadPath(flowRequest.ProjectID, flowRequest.DirectoryID, flowRequest.FileID),
	}
}
