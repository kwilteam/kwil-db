package rx

import (
	"context"
	"errors"
	"unsafe"
)

// ErrCancelled is the error returned when a Task/Promise/Continuation
// has been previously cancelled
var ErrCancelled = errors.New("cancelled prior to completion")

// Task is the primary implementation of the Promise interface
// It can be used as the controlling handle for setting completion
// of a Promise or as a Promise itself.
type Task[T any] struct {
	status *uint32
	state  unsafe.Pointer
}

// GetError will return the contained error or nil if the
// result is not an error
// NOTE: this is a blocking call
func (r *Task[T]) GetError() error { return r._getError() }

// Get will panic if the Result is in an errored state
// otherwise it will return the contained value or the default
// value of the underlying type if it is nil
// NOTE: this is a blocking call
func (r *Task[T]) Get() T { return r._get() }

// IsError will return true if the Result is an error.
// It will return false if it has not yet completed.
func (r *Task[T]) IsError() bool { return r._isError() }

// IsCancelled will return true if the Result is cancelled.
// It will return false if it has not yet completed.
func (r *Task[T]) IsCancelled() bool { return r._isCancelled() }

func (r *Task[T]) IsErrorOrCancelled() bool { return r._isErrorOrCancelled() }

// IsDone will return true if the Result is complete
func (r *Task[T]) IsDone() bool { return r._isDone() }

// DoneChan will return a channel that will be closed when the
// result/error has been set
func (r *Task[T]) DoneChan() <-chan struct{} { return r._doneChan() }

// Fail will set the result to an error
func (r *Task[T]) Fail(err error) bool { return r._fail(err) }

// Complete will set the result to a value
func (r *Task[T]) Complete(value T) bool { return r._complete(value) }

// CompleteOrFail will set the result to either a value or an error
func (r *Task[T]) CompleteOrFail(value T, err error) bool { return r._completeOrFail(value, err) }

// Cancel will cancel the Continuation
func (r *Task[T]) Cancel() bool { return r._cancel() }

// GetOrError will return the error with the value
// If the value of the underlying type if it is nil,
// then it will return the default value of the
// underlying type
// NOTE: this is a blocking call
func (r *Task[T]) GetOrError() (T, error) { return r._getOrError() }

// Await will block until the result is complete or the context
// is cancelled, reached its timeout or deadline. 'ok' will be true
// if the result is complete, otherwise it will be false. Passing a
// nil ctx will block until result completion.
// NOTE: this is a blocking call
func (r *Task[T]) Await(ctx context.Context) (ok bool) { return r._await(ctx) }

// Then will call the func when the result has been successfully set
func (r *Task[T]) Then(fn ValueHandler[T]) *Task[T] { return r._then(fn) }

// Catch will call the func if the result is an error
func (r *Task[T]) Catch(fn ErrorHandler) *Task[T] { return r._catch(fn) }

// WhenComplete will call the func when the result has been set
func (r *Task[T]) WhenComplete(fn Handler[T]) *Task[T] { return r._whenComplete(fn) }

// OnCompleteRun will call the func when the result has been set
func (r *Task[T]) OnCompleteRun(fn Runnable) { r._onCompleteRun(fn) }

// OnComplete will call the func when the result has been set
func (r *Task[T]) OnComplete(fn Handler[T]) { r._onComplete(fn) }

// ThenAsync will asynchronously call the func when the result has
// been successfully set
func (r *Task[T]) ThenAsync(fn ValueHandler[T]) *Task[T] { return r._thenAsync(fn) }

// CatchAsync will asynchronously call the func if the
// result is an error
func (r *Task[T]) CatchAsync(fn ErrorHandler) *Task[T] { return r._catchErrorAsync(fn) }

// WhenCompleteAsync will asynchronously call the func when the result
// has been set
func (r *Task[T]) WhenCompleteAsync(fn Handler[T]) *Task[T] { return r._whenCompleteAsync(fn) }

// OnCompleteAsync will asynchronously call the func when the result
// has been set
func (r *Task[T]) OnCompleteAsync(fn Handler[T]) { r._onCompleteAsync(fn) }

// OnCompleteRunAsync will asynchronously call the func when the result
// has been set
func (r *Task[T]) OnCompleteRunAsync(fn Runnable) { r._onCompleteRunAsync(fn) }

// AsContinuation returns a continuation that will be completed
// when the source task has been completed
func (r *Task[T]) AsContinuation() *Continuation {
	return r._asContinuation(false)
}

// AsContinuationAsync returns a continuation that will be completed
// asynchronously when the source task has been completed
func (r *Task[T]) AsContinuationAsync() *Continuation {
	return r._asContinuation(true)
}

// AsAsync returns a task that will be completed asynchronously
// when the source task has been completed
func (r *Task[T]) AsAsync() *Task[T] {
	return r._async()
}
