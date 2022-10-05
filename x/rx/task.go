package rx

import (
	"context"
	"errors"
)

// ErrCancelled is the error returned when a task/Continuation
// has been previously cancelled
var ErrCancelled = errors.New("cancelled prior to completion")

// Void An empty struct{} typed as 'Void'
type Void struct{}

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
	DoneChan() <-chan Void

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

	// WhenComplete will call the func when the result has been set
	WhenComplete(fn Handler[T]) Task[T]

	// OnCompleteRun will call the func when the result has been set
	OnCompleteRun(fn Runnable)

	// OnComplete will call the func when the result has been set
	OnComplete(fn Handler[T])

	// ThenAsync will asynchronously call the func when the result has
	// been successfully set
	ThenAsync(fn ValueHandler[T]) Task[T]

	// CatchAsync will asynchronously call the func if the
	// result is an error
	CatchAsync(fn ErrorHandler) Task[T]

	// WhenCompleteAsync will asynchronously call the func when the result
	// has been set
	WhenCompleteAsync(fn Handler[T]) Task[T]

	// OnCompleteAsync will asynchronously call the func when the result
	// has been set
	OnCompleteAsync(fn Handler[T])

	// OnCompleteRunAsync will asynchronously call the func when the result
	// has been set
	OnCompleteRunAsync(fn Runnable)

	// AsContinuation returns a continuation that will be completed
	// when the source task has been completed
	AsContinuation() *Continuation

	// AsContinuationAsync returns a continuation that will be completed
	// asynchronously when the source task has been completed
	AsContinuationAsync() *Continuation

	// AsAsync returns a task that will be completed asynchronously
	// when the source task has been completed
	AsAsync() Task[T]

	// IsAsync returns true if the task was created to call
	// func continuations asynchronously
	IsAsync() bool
}
