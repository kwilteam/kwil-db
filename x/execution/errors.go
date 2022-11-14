package execution

import "errors"

var ErrUnauthorized = errors.New("unauthorized to execute query")
var ErrQueryNotFound = errors.New("query not found")
var ErrRoleNotFound = errors.New("role not found")
var ErrRoleDoesNotHavePermission = errors.New("role does not have permission")
var ErrUnknownType = errors.New("unknown type")
