package x

import "unsafe"

type Result[T any] struct {
	state uint8
	value unsafe.Pointer
}

// ResultEmptySuccess will return a new Result in a successful state
func ResultEmptySuccess[T any]() *Result[T] {
	return &Result[T]{state: 3}
}

// ResultSuccess will return a new Result in a successful state
func ResultSuccess[T any](value T) *Result[T] {
	return &Result[T]{state: 1, value: unsafe.Pointer(&value)}
}

// ResultFailure will return a new Result in an errored state
func ResultFailure[T any](err error) *Result[T] {
	return &Result[T]{state: 2, value: unsafe.Pointer(&err)}
}

// GetError will return the contained error or nil if the
// result is not an error
func (r *Result[T]) GetError() error {
	if r.IsError() {
		return *(*error)(r.value)
	}
	return nil
}

// Get will panic if the Result is in an errored state
// otherwise it will return the contained value or the default
// value of the underlying type if it is nil
func (r *Result[T]) Get() (value T) {
	if r.state == 1 {
		return *(*T)(r.value)
	}

	if !r.IsError() {
		return
	}

	panic(*(*error)(r.value))
}

// GetOrError will return the error with the value
// If the value of the underlying type if it is nil,
// then it will return the default value of the
// underlying type
func (r *Result[T]) GetOrError() (value T, err error) {
	if r.state == 1 {
		value = *(*T)(r.value)
		return
	}

	if r.IsError() {
		err = *(*error)(r.value)
		return
	}

	return
}

// GetOrDefault will panic if the Result is in an errored
// state. Otherwise, it will return the contained value or the
// otherwise it will return the contained value or the alt
// value if the underlying type is nil or not set
func (r *Result[T]) GetOrDefault(alt T) (value T) {
	if r.state == 1 {
		return *(*T)(r.value)
	}

	if r.state == 0 || r.state == 3 {
		return alt
	}

	panic(*(*error)(r.value))
}

// IsError will return true if the Result is an error
func (r *Result[T]) IsError() bool {
	return r.state == 2
}

// IsSet will return true if the Result is set
func (r *Result[T]) IsSet() bool {
	return r.state != 0
}

// IsNilOrDefault will return true if the Result is set to
// the default value for type T (e.g., nil, etc)
func (r *Result[T]) IsNilOrDefault() bool {
	return r.state == 3
}
