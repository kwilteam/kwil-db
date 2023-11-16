package registry

import (
	"errors"
)

var (
	ErrDatabaseExists         = errors.New("database already exists")
	ErrDatabaseNotFound       = errors.New("database not found")
	ErrStillDeploying         = errors.New("database has not finished deploying")
	ErrRegistryNotWritable    = errors.New("registry is not writable")
	ErrAlreadyInSession       = errors.New("already in session")
	ErrWritable               = errors.New("registry is writable at an unexpected time")
	ErrIdempotencyKeyMismatch = errors.New("idempotency key mismatch")
)
