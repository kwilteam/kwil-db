package execution

import "errors"

var (
	ErrUnitializedExtension     = errors.New("extension not initialized")
	ErrUnknownExtension         = errors.New("unknown extension")
	ErrUnknownPreparedStatement = errors.New("unknown prepared statement")
	ErrUnknownVariable          = errors.New("unknown variable")
	ErrAccessControl            = errors.New("access control failed")
	ErrUnknownProcedure         = errors.New("unknown procedure")
	ErrIncorrectNumArgs         = errors.New("incorrect number of arguments")
	ErrIncorrectInputType       = errors.New("incorrect input type")
	ErrScopingViolation         = errors.New("scoping violation")
	ErrMutativeStatement        = errors.New("mutative statement")
)
