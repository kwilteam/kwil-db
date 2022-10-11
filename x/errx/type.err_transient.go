package errx

import (
	"errors"
	"fmt"
)

// ErrTransient to be used to represent an error
// that is only temporary.
type ErrTransient interface {
	Error() string
	Unwrap() error

	unwrapTransient() error
}

type transientError struct {
	err error
}

func (e transientError) Unwrap() error {
	if e.err == nil {
		return errors.New(e.getTransientErrorMessage())
	}

	return e.unwrapTransient()
}

func (e transientError) Error() string {
	return e.err.Error()
}

func (e transientError) unwrapTransient() error {
	if e.err == nil {
		return fmt.Errorf(e.getTransientErrorMessage())
	}
	return e.err
}

func (e transientError) getTransientErrorMessage() string {
	return fmt.Sprintf("transient error")
}
