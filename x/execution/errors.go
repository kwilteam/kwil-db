package execution

import "errors"

var ErrUnauthorized = errors.New("unauthorized to execute query")
var ErrQueryNotFound = errors.New("query not found")
