package driver

import "errors"

var (
	ErrConnectionClosed  = errors.New("connection closed")
	ErrNoWriteLock       = errors.New("no write lock")
	ErrActiveSavepoint   = errors.New("savepoint already active")
	ErrSavepointRollback = errors.New("savepoint rollback")
	ErrNoActiveSavepoint = errors.New("no active savepoint")
	ErrLockWaitTimeout   = errors.New("lock wait timeout")
)
