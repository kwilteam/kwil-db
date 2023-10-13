package dataset

import "errors"

var (
	ErrExtensionNotFound      = errors.New("extension not found")
	ErrCallerNotAuthenticated = errors.New("caller not authenticated")
	ErrCallerNotOwner         = errors.New("caller not owner")
	ErrCallMutativeProcedure  = errors.New("cannot call mutative procedure")
)
