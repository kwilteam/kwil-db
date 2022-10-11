package rx

import (
	"context"
	. "kwil/x"
)

// Action is for use with non typed functional
// continuations from tasks
type Action interface {
	// Fail will set the result to an error
	Fail(err error) bool

	// Complete will set the result to a value
	Complete() bool

	// CompleteOrFail will set the result to either a value or an error
	CompleteOrFail(err error) bool

	// Cancel will cancel the Action
	Cancel() bool

	// IsDone will return true if the Result is complete
	IsDone() bool

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

	// Await will block until the result is complete or the context
	// is cancelled, reached its timeout or deadline. 'ok' will be true
	// if the result is complete, otherwise it will be false. Passing a
	// nil ctx will block until result completion.
	Await(ctx context.Context) bool

	// GetError will return the contained error or nil if the
	// result is not an error
	// NOTE: this is a blocking call
	GetError() error

	// DoneChan will return a channel that will be closed when the
	// result/error has been set
	DoneChan() <-chan Void

	// Then will call the func when the result has been successfully set
	Then(fn func()) Action

	// Catch will call the func if the result is an error
	Catch(fn func(error)) Action

	// ThenCatchFinally will call the func when the result has been set
	ThenCatchFinally(fn *ContinuationA) Action

	// WhenComplete will call the func when the result has been set
	WhenComplete(fn func(error)) Action

	// OnComplete will call the func when the result has been set
	OnComplete(*ContinuationT[Void])

	// AsAction returns an opaque continuation that will be completed
	// when the source task has been completed
	AsAction() Action

	// AsListenable is a convenience method for casting the current Task
	// to a Listenable
	AsListenable() Listenable[Void]

	// AsAsync returns a continuation that will be completed asynchronously
	// when the source task has been completed, using the provided
	// executor. If the executor is nil, then the default async executor
	// will be used.
	AsAsync(e Executor) Action

	// IsAsync returns true if the continuation was created to call
	// continuations asynchronously
	IsAsync() bool
}
