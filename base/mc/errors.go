package mc

import (
	"errors"
)

var (
	// ErrNotFound Item not found
	ErrNotFound = errors.New("not found")

	// ErrInvalid Invalid request
	ErrInvalid = errors.New("invalid")

	// ErrExists Items already exists
	ErrExists = errors.New("exists")

	// ErrNoAccess Access to item not allowed
	ErrNoAccess = errors.New("no access")

	// ErrCreate Create failed reason unknown
	ErrCreate = errors.New("unable to create")

	// ErrInternal Internal fatal error
	ErrInternal = errors.New("internal error")

	// ErrInUse object is locked and in use by someone else
	ErrInUse = errors.New("in use")
)

// ErrorCode is an integer representation of a error that we can encode and send
// over the network.
type ErrorCode int

const (
	// ErrorCodeSuccess err == nil
	ErrorCodeSuccess ErrorCode = iota

	// ErrorCodeNotFound ErrNotFound
	ErrorCodeNotFound

	// ErrorCodeInvalid ErrInvalid
	ErrorCodeInvalid

	// ErrorCodeExists ErrCodeExists
	ErrorCodeExists

	// ErrorCodeNoAccess ErrNoAccess
	ErrorCodeNoAccess

	// ErrorCodeCreate ErrCreate
	ErrorCodeCreate

	// ErrorCodeInternal ErrInternal
	ErrorCodeInternal

	// ErrorCodeInUse ErrInUse
	ErrorCodeInUse

	// ErrorCodeUnknown Catch all when we can't map an error
	ErrorCodeUnknown
)

var errorCodeMapping = map[ErrorCode]error{
	ErrorCodeSuccess:  nil,
	ErrorCodeNotFound: ErrNotFound,
	ErrorCodeInvalid:  ErrInvalid,
	ErrorCodeExists:   ErrExists,
	ErrorCodeNoAccess: ErrNoAccess,
	ErrorCodeCreate:   ErrCreate,
	ErrorCodeInternal: ErrInternal,
	ErrorCodeInUse:    ErrInUse,
}

// ErrorCodeToError maps a given ErrorCode to an error.
func ErrorCodeToError(ec ErrorCode) error {
	return errorCodeMapping[ec]
}

var errorMapping = map[string]ErrorCode{
	ErrNotFound.Error(): ErrorCodeNotFound,
	ErrInvalid.Error():  ErrorCodeInvalid,
	ErrExists.Error():   ErrorCodeExists,
	ErrNoAccess.Error(): ErrorCodeNoAccess,
	ErrCreate.Error():   ErrorCodeCreate,
	ErrInternal.Error(): ErrorCodeInternal,
	ErrInUse.Error():    ErrorCodeInUse,
}

// ErrorToErrorCode maps from an error to an ErrorCode.
func ErrorToErrorCode(err error) ErrorCode {
	return errorMapping[err.Error()]
}
