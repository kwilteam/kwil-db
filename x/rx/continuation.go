package rx

import (
	"context"
	"kwil/x"
)

// Continuation is for use with non typed functional
// continuations from tasks
type Continuation interface {

	// Fail will set the result to an error
	Fail(err error) bool

	// Complete will set the result to a value
	Complete() bool

	// CompleteOrFail will set the result to either a value or an error
	CompleteOrFail(err error) bool

	// Cancel will cancel the Continuation
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
	DoneChan() <-chan x.Void

	// Then will call the func when the result has been successfully set
	Then(fn Runnable) Continuation

	// Catch will call the func if the result is an error
	Catch(fn ErrorHandler) Continuation

	// ThenCatchFinally will call the func when the result has been set
	ThenCatchFinally(fn *CompletionC) Continuation

	// WhenComplete will call the func when the result has been set
	WhenComplete(fn func(error)) Continuation

	// OnComplete will call the func when the result has been set
	OnComplete(fn *Completion[x.Void])

	// AsContinuation returns an opaque continuation that will be completed
	// when the source task has been completed
	AsContinuation() Continuation

	// AsAsync returns an opaque continuation that will be completed
	// asynchronously when the source continuation has been
	// completed
	AsAsync() Continuation

	// IsAsync returns true if the continuation was created to call
	// continuations asynchronously
	IsAsync() bool
}
