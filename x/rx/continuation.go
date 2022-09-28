package rx

import "context"

// Continuation is for us with non typed functional continuations
// from tasks/promises
type Continuation struct {
	task *Task[struct{}]
}

// Cancel will cancel the Continuation
func (c *Continuation) Cancel() bool { return c.task.Cancel() }

// IsDone will return true if the Result is complete
func (c *Continuation) IsDone() bool {
	return c.task.IsDone()
}

// IsError will return true if the Result is an erroc.
// It will return false if it has not yet completed.
func (c *Continuation) IsError() bool {
	return c.task.IsError()
}

// IsCancelled will return true if the Result is cancelled.
// It will return false if it has not yet completed.
func (c *Continuation) IsCancelled() bool {
	return c.task.IsCancelled()
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
//
// NOTE: this is a blocking call
func (c *Continuation) GetError() error {
	return c.task.GetError()
}

// DoneChan will return a channel that will be closed when the
// result/error has been set
func (c *Continuation) DoneChan() <-chan struct{} {
	return c.task.DoneChan()
}

// Then will call the func when the result has been successfully set
func (c *Continuation) Then(fn Runnable) *Continuation {
	h := onSuccessContinuationHandler{fn: fn}
	return &Continuation{task: c.task.WhenComplete(h.invoke)}
}

// Catch will call the func if the result is an error
func (c *Continuation) Catch(fn ErrorHandler) *Continuation {
	return &Continuation{task: c.task.Catch(fn)}
}

// OnComplete will call the func when the result has been set
func (c *Continuation) OnComplete(fn func(error)) *Continuation {
	r := &continuationRequestAdapter{fn: fn}
	return &Continuation{task: c.task.WhenComplete(r.invoke)}
}

// WhenComplete will call the func when the result has been set
func (c *Continuation) WhenComplete(fn Handler[struct{}]) {
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
	return &Continuation{task: c.task.WhenComplete(h.invoke)}
}

// CatchAsync will asynchronously call the func if the
// result is an error
func (c *Continuation) CatchAsync(fn ErrorHandler) *Continuation {
	return &Continuation{task: c.task.CatchAsync(fn)}
}

// WhenCompleteAsync will asynchronously call the func when the result
// has been set
func (c *Continuation) WhenCompleteAsync(fn func(error)) *Continuation {
	r := &continuationRequestAdapterAsync{fn: fn}
	return &Continuation{task: c.task.WhenComplete(r.invoke)}
}

// OnCompleteAsync will asynchronously call the func when the result
// has been set
func (c *Continuation) OnCompleteAsync(fn Handler[struct{}]) {
	c.task.OnCompleteAsync(fn)
}

// OnCompleteRunAsync will asynchronously call the func when the result
// has been set
func (c *Continuation) OnCompleteRunAsync(fn Runnable) {
	c.task.OnCompleteRunAsync(fn)
}

// AsAsync returns a continuation that will be completed
// asynchronously when the source continuation has been
// completed
func (c *Continuation) AsAsync() *Continuation {
	return c.task.AsContinuationAsync()
}
