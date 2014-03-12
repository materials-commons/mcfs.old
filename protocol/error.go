package protocol

import (
	"fmt"
	"github.com/materials-commons/base/mc"
)

type MCFSError struct {
	ErrorMessage string
	Err          error
}

func (e *MCFSError) Error() string {
	return e.Err.Error() + ":" + e.ErrorMessage
}

func (e *MCFSError) ToErrorCode() mc.ErrorCode {
	return mc.ErrorToErrorCode(e.Err)
}

func FromErrorCode(errorCode mc.ErrorCode) *MCFSError {
	return &MCFSError{
		Err: mc.ErrorCodeToError(errorCode),
	}
}

func NewMCFSError(err error, msg string) *MCFSError {
	return &MCFSError{
		ErrorMessage: msg,
		Err:          err,
	}
}

func Errorf(err error, message string, args ...interface{}) *MCFSError {
	msg := fmt.Sprintf(message, args...)
	return NewMCFSError(err, msg)
}
