package sqlx

import "errors"

var (
	ErrLocked = errors.New("sql/schema: lock is held by other session")
)
