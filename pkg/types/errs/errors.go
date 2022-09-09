package errs

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")

var ErrDBExists = errors.New("database already exists")

var ErrDBNotFound = errors.New("database not found")
