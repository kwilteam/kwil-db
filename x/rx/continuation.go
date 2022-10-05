package rx

import "context"

// Continuation is for use with non typed functional
// continuations from tasks
type Continuation struct {
	task *task[Void]
}

// Fail will set the result to an error
func (c *Continuation) Fail(err error) bool { return c.task.Fail(err) }

// Complete will set the result to a value
func (c *Continuation) Complete() bool { return c.task.Complete(Void{}) }

// CompleteOrFail will set the result to either a value or an error
func (c *Continuation) CompleteOrFail(err error) bool { return c.task.CompleteOrFail(Void{}, err) }

// Cancel will cancel the Continuation
func (c *Continuation) Cancel() bool { return c.task.Cancel() }

// IsDone will return true if the Result is complete
func (c *Continuation) IsDone() bool {
	return c.task.IsDone()
}

// IsError will return true if the Result is an error.
// It will return false if it has not yet completed.
func (c *Continuation) IsError() bool {
	return c.task.IsError()
}

// IsCancelled will return true if the Result is cancelled.
// It will return false if it has not yet completed.
func (c *Continuation) IsCancelled() bool {
	return c.task.IsCancelled()
}

// IsErrorOrCancelled will return true if the Result is an error
// or cancelled (NOTE: cancelled is always an error).
// It will return false if it has not yet completed.
func (c *Continuation) IsErrorOrCancelled() bool {
	return c.task.IsErrorOrCancelled()
}

// Await will block until the result is complete or the context
// is cancelled, reached its timeout or deadline. 'ok' will be true
// if the result is complete, otherwise it will be false. Passing a
// nil ctx will block until result completion.
func (c *Continuation) Await(ctx context.Context) bool {
	return c.task.Await(ctx)
}

// GetError will return the contained error or nil if the
// result is not an error
// NOTE: this is a blocking call
func (c *Continuation) GetError() error {
	return c.task.GetError()
}

// DoneChan will return a channel that will be closed when the
// result/error has been set
func (c *Continuation) DoneChan() <-chan Void {
	return c.task.DoneChan()
}

// Then will call the func when the result has been successfully set
func (c *Continuation) Then(fn Runnable) *Continuation {
	h := onSuccessContinuationHandler{fn: fn}
	return c.task.WhenComplete(h.invoke).AsContinuation()
}

// Catch will call the func if the result is an error
func (c *Continuation) Catch(fn ErrorHandler) *Continuation {
	return c.task.Catch(fn).AsContinuation()
}

// OnComplete will call the func when the result has been set
func (c *Continuation) OnComplete(fn func(error)) *Continuation {
	r := &continuationRequestAdapter{fn: fn}
	return c.task.WhenComplete(r.invoke).AsContinuation()
}

// WhenComplete will call the func when the result has been set
func (c *Continuation) WhenComplete(fn Handler[Void]) {
	c.task.WhenComplete(fn)
}

// OnCompleteRun will call the func when the result has been set
func (c *Continuation) OnCompleteRun(fn Runnable) {
	c.task.OnCompleteRun(fn)
}

// ThenAsync will asynchronously call the func when the result has
// been successfully set
func (c *Continuation) ThenAsync(fn Runnable) *Continuation {
	h := onSuccessContinuationHandlerAsync{fn: fn}
	return c.task.WhenComplete(h.invoke).AsContinuation()
}

// CatchAsync will asynchronously call the func if the
// result is an error
func (c *Continuation) CatchAsync(fn ErrorHandler) *Continuation {
	return c.task.CatchAsync(fn).AsContinuation()
}

// WhenCompleteAsync will asynchronously call the func when the result
// has been set
func (c *Continuation) WhenCompleteAsync(fn func(error)) *Continuation {
	r := &continuationRequestAdapterAsync{fn: fn}
	return c.task.WhenComplete(r.invoke).AsContinuation()
}

// OnCompleteAsync will asynchronously call the func when the result
// has been set
func (c *Continuation) OnCompleteAsync(fn Handler[Void]) {
	c.task.OnCompleteAsync(fn)
}

// OnCompleteRunAsync will asynchronously call the func when the result
// has been set
func (c *Continuation) OnCompleteRunAsync(fn Runnable) {
	c.task.OnCompleteRunAsync(fn)
}

// AsContinuation returns an opaque continuation that will be completed
// when the source task has been completed
func (c *Continuation) AsContinuation() *Continuation {
	return c.task.AsContinuation()
}

// AsAsync returns an opaque continuation that will be completed
// asynchronously when the source continuation has been
// completed
func (c *Continuation) AsAsync() *Continuation {
	return c.task.AsContinuationAsync()
}

// IsAsync returns true if the continuation was created to call
// continuations asynchronously
func (c *Continuation) IsAsync() bool {
	return c.task.IsAsync()
}
