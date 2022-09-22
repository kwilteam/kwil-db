package schema

import "errors"

// ErrLocked is returned on Lock calls which have failed to obtain the lock.
var ErrLocked = errors.New("sql/schema: lock is held by other session")
