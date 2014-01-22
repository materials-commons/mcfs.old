package mcfs

import (
	"errors"
)

var (
	ErrNotExist       = errors.New("Does not exist")
	ErrInvalidRequest = errors.New("Invalid Request")
)
