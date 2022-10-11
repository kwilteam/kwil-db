package errx

import (
	"context"
	"errors"
	"time"
)

var _errOperationCancelled = errors.New("operation cancelled prior to completion")

// ErrOperationCancelled to be used to return for an operation that
// has been cancelled prior to completion.
func ErrOperationCancelled() error {
	return _errOperationCancelled
}

// Transient returns an error that implements the ErrTransient interface.
// A default error message will be returned if nil.
func Transient(err error) ErrTransient {
	return &transientError{err: err}
}

// RateLimited returns an error that implements the ErrRateLimited interface.
// A default error message will be returned if nil. A negative retryAfter value
// will be converted to a zero duration.
func RateLimited(err error, retryAfter time.Duration) ErrRateLimited {
	if retryAfter < 0 {
		retryAfter = 0
	}

	return &rateLimitedError{
		err:        err,
		retryAfter: retryAfter,
	}
}

// IsOperationCancelled returns true if the underlying error is
// ErrOperationCancelled.
func IsOperationCancelled(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, _errOperationCancelled) || _isContextTimeout(err)
}

// IsCancelled returns true if the underlying error is ErrOperationCancelled
// or a context timeout as returned true by IsContextTimeout.
func IsCancelled(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, _errOperationCancelled) || _isContextTimeout(err)
}

// IsTransient returns true if the underlying error is ErrTransient
func IsTransient(err error) bool {
	_, ok := err.(ErrTransient)
	return ok
}

// IsRateLimited is true if the underlying error is ErrRateLimited
func IsRateLimited(err error) bool {
	_, ok := err.(ErrRateLimited)
	return ok
}

// IsContextTimeout returns true if the underlying error is a timeout.
// This function counts canceled contexts as timeouts.
// ref: https://blog.afoolishmanifesto.com/posts/context-deadlines-in-golang/
func IsContextTimeout(err error) bool {
	return err != nil && _isContextTimeout(err)
}

func _isContextTimeout(err error) bool {
	switch {
	case errors.Is(err, context.Canceled):
		return true
	case errors.Is(err, context.DeadlineExceeded):
		return true
	default:
		return false
	}
}
