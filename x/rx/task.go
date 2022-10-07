package rx

import (
	"context"
	"kwil/x"
)

// Task is a Promise-like interface. It can be used as the
// controller for setting completion of a Continuation or
// a task itself.
type Task[T any] interface {
	// GetError will return the contained error or nil if the
	// result is not an error
	// NOTE: this is a blocking call
	GetError() error

	// Get will panic if the Result is in an errored state
	// otherwise it will return the contained value or the default
	// value of the underlying type if it is nil
	// NOTE: this is a blocking call
	Get() T

	// GetOrError will return the error with the value
	// If the value of the underlying type if it is nil,
	// then it will return the default value of the
	// underlying type
	// NOTE: this is a blocking call
	GetOrError() (T, error)

	// IsError will return true if the Result is an error.
	// It will return false if it has not yet completed.
	IsError() bool

	// IsCancelled will return true if the Result is cancelled.
	// It will return false if it has not yet completed.
	IsCancelled() bool

	// IsErrorOrCancelled will return true if the Result is an error
	// or cancelled (NOTE: cancelled is always an error).
	// It will return false if it has not yet completed.
	IsErrorOrCancelled() bool

	// IsDone will return true if the Result is complete
	IsDone() bool

	// DoneChan will return a channel that will be closed when the
	// result/error has been set
	DoneChan() <-chan x.Void

	// Await will block until the result is complete or the context
	// is cancelled, reached its timeout or deadline. 'ok' will be true
	// if the result is complete, otherwise it will be false. Passing a
	// nil ctx will block until result completion.
	// NOTE: this is a blocking call
	Await(ctx context.Context) (ok bool)

	// Fail will set the result to an error
	Fail(err error) bool

	// Complete will set the result to a value
	Complete(value T) bool

	// CompleteOrFail will set the result to either a value or an error
	CompleteOrFail(value T, err error) bool

	// Cancel will cancel the task
	Cancel() bool

	// Then will call the func when the result has been successfully set
	Then(fn ValueHandler[T]) Task[T]

	// Catch will call the func if the result is an error
	Catch(fn ErrorHandler) Task[T]

	// Handle will call the func when the result has been set
	Handle(fn func(T, error) (T, error)) Task[T]

	// Compose will call the func when the result has been set
	Compose(fn func(T, error) Task[T]) Task[T]

	// ThenCatchFinally will call the func when the result has been set
	ThenCatchFinally(fn *Completion[T]) Task[T]

	// WhenComplete will call the func when the result has been set
	WhenComplete(fn Handler[T]) Task[T]

	// OnComplete will call the func when the result has been set
	OnComplete(fn *Completion[T])

	// AsContinuation returns a continuation that will be completed
	// when the source task has been completed
	AsContinuation() Continuation

	// AsContinuationAsync returns a continuation that will be completed
	// asynchronously when the source task has been completed
	AsContinuationAsync() Continuation

	// AsAsync returns a task that will be completed asynchronously
	// when the source task has been completed
	AsAsync() Task[T]

	// IsAsync returns true if the task was created to call
	// func continuations asynchronously
	IsAsync() bool
}
