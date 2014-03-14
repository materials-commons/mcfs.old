package protocol

import (
	"fmt"
	"github.com/materials-commons/base/mc"
)

type MCFSError struct {
	Err     error
	Message string
}

func (e *MCFSError) Error() string {
	return e.Err.Error() + ":" + e.Message
}

func (e *MCFSError) ToErrorCode() mc.ErrorCode {
	return mc.ErrorToErrorCode(e.Err)
}

func FromErrorCode(errorCode mc.ErrorCode) *MCFSError {
	return &MCFSError{
		Err: mc.ErrorCodeToError(errorCode),
	}
}

func newMCFSError(err error, msg string) *MCFSError {
	return &MCFSError{
		Message: msg,
		Err:     err,
	}
}

func Errorf(err error, message string, args ...interface{}) *MCFSError {
	msg := fmt.Sprintf(message, args...)
	return newMCFSError(err, msg)
}
