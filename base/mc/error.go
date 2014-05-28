package mc

import (
	"fmt"
)

// Error holds the error code and additional messages.
type Error struct {
	Err     error
	Message string
}

// Implement error interface.
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Err.Error() + ":" + e.Message
	}

	return e.Err.Error()
}

// ToErrorCode converts an Error to an ErrorCode
func (e *Error) ToErrorCode() ErrorCode {
	return ErrorToErrorCode(e.Err)
}

// FromErrorCode takes an error code and returns the corresponding Error.
func FromErrorCode(errorCode ErrorCode) *Error {
	return &Error{
		Err: ErrorCodeToError(errorCode),
	}
}

// newError creates a new instance of an Error.
func newError(err error, msg string) *Error {
	return &Error{
		Message: msg,
		Err:     err,
	}
}

func Is(err error, what error) bool {
	if e, ok := err.(*Error); ok {
		return e.Err == what
	}

	return err == what
}

// Errorf takes and error, a message string and a set of arguments and produces
// a new Error.
func Errorf(err error, message string, args ...interface{}) *Error {
	msg := fmt.Sprintf(message, args...)
	return newError(err, msg)
}

// Errorm takes an mc error code, and another error, treats the other error
// as a status message to put in the Message field.
func Errorm(err error, err2 error) *Error {
	return newError(err, err2.Error())
}
