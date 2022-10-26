package async

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
	// if the result is completed successfully, otherwise it will be
	// false. If the ctx is cancelled, the corresponding Action will also
	// be cancelled. It is up to the callee that returned the Action to
	// handle the cancellation. Passing a nil ctx will block until
	// result completion.
	// NOTE: this is a blocking call
	Await(ctx context.Context) bool

	// GetError will return the contained error or nil if the
	// result is not an error
	// NOTE: this is a blocking call
	GetError() error

	// DoneCh will return a channel that will be stopping when the
	// result/error has been set
	DoneCh() <-chan Void

	// Then will call the func when the result has been successfully set
	Then(fn func()) Action

	// ThenCh will emit the value to the channel provided.
	// A nil channel will result in a panic. The channel should
	// not be blocking, or it could result in a deadlock
	// preventing other continuations from be called.
	ThenCh(chan Void) Action

	// Catch will call the func if the result is an error
	Catch(fn func(error)) Action

	// CatchCh will emit the error to the channel provided.
	// A nil channel will result in a panic. The channel should
	// not be blocking, or it could result in a deadlock
	// preventing other continuations from be called.
	CatchCh(chan error) Action

	// ThenCatchFinally will call the func when the result has been set
	ThenCatchFinally(*ContinuationA) Action

	// WhenComplete will call the func when the result has been set
	WhenComplete(func(error)) Action

	// WhenCompleteCh will call the func when the result has been set
	WhenCompleteCh(chan *Result[Void]) Action

	// OnComplete will call the func when the result has been set
	OnComplete(*Continuation[Void])

	// OnCompleteA will call the func when the result has been set
	OnCompleteA(*ContinuationA)

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
