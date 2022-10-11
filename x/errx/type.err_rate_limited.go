package errx

import (
	"fmt"
	"time"
)

// ErrRateLimited to be used to return for an operation that
// has been rate limited and can be retried.
type ErrRateLimited interface {
	RetryAfter() time.Duration
	Error() string
	Unwrap() error

	unwrapRateLimited() error
	unwrapTransient() error
}

type rateLimitedError struct {
	err        error
	retryAfter time.Duration
}

func (e rateLimitedError) Unwrap() error {
	return e.unwrapRateLimited()
}

func (e rateLimitedError) Error() string {
	if e.err == nil {
		return e.getRateLimitDefaultErrorMessage()
	}

	return e.err.Error()
}

func (e rateLimitedError) RetryAfter() time.Duration {
	return e.retryAfter
}

func (e rateLimitedError) unwrapRateLimited() error {
	if e.err == nil {
		return fmt.Errorf(e.getRateLimitDefaultErrorMessage())
	}
	return e.err
}

func (e rateLimitedError) unwrapTransient() error {
	return e.unwrapRateLimited()
}

func (e rateLimitedError) getRateLimitDefaultErrorMessage() string {
	return fmt.Sprintf("rate limited, retry after %s", e.retryAfter)
}
