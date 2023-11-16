package sqlite

import "errors"

var (
	ErrWriterOpen    = errors.New("writer connection is already open")
	ErrClosed        = errors.New("result set is closed")
	ErrFloatDetected = errors.New("float detected")
	ErrReadOnlyConn  = errors.New("connection is read only")
	ErrInUse         = errors.New("connection is in use")
	ErrInterrupted   = errors.New("execution was interrupted")
)
