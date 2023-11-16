package sessions

import "errors"

var (
	ErrInSession              = errors.New("cannot begin session, already in a session")
	ErrNotInSession           = errors.New("cannot commit, not in a session")
	ErrIdempotencyKeyMismatch = errors.New("idempotency key mismatch")
)
