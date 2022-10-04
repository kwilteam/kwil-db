package core

import "errors"

var (
	ErrUnsupportedOS = errors.New("the PostgreSQL engine does not support Windows")
)
