package x

import "errors"

// ErrOperationCancelled is an error returned when an operation
// has been cancelled prior to completion
var ErrOperationCancelled = errors.New("operation cancelled prior to completion")
