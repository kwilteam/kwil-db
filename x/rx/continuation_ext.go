package rx

import (
	"context"
	"kwil/x"
)

type continuation struct {
	task *task[x.Void]
}

func (c *continuation) Fail(err error) bool { return c.task.Fail(err) }

func (c *continuation) Complete() bool { return c.task.Complete(x.Void{}) }

func (c *continuation) CompleteOrFail(err error) bool { return c.task.CompleteOrFail(x.Void{}, err) }

func (c *continuation) Cancel() bool { return c.task.Cancel() }

func (c *continuation) IsDone() bool {
	return c.task.IsDone()
}

func (c *continuation) IsError() bool {
	return c.task.IsError()
}

func (c *continuation) IsCancelled() bool {
	return c.task.IsCancelled()
}

func (c *continuation) IsErrorOrCancelled() bool {
	return c.task.IsErrorOrCancelled()
}

func (c *continuation) Await(ctx context.Context) bool {
	return c.task.Await(ctx)
}

func (c *continuation) GetError() error {
	return c.task.GetError()
}

func (c *continuation) DoneChan() <-chan x.Void {
	return c.task.DoneChan()
}

func (c *continuation) Then(fn Runnable) Continuation {
	h := onSuccessContinuationHandler{fn: fn}
	return c.task.WhenComplete(h.invoke).AsContinuation()
}

func (c *continuation) Catch(fn ErrorHandler) Continuation {
	return c.task.Catch(fn).AsContinuation()
}

func (c *continuation) OnComplete(fn func(error)) Continuation {
	r := &continuationRequestAdapter{fn: fn}
	return c.task.WhenComplete(r.invoke).AsContinuation()
}

func (c *continuation) WhenComplete(fn Handler[x.Void]) {
	c.task.WhenComplete(fn)
}

func (c *continuation) OnCompleteRun(fn Runnable) {
	c.task.OnCompleteRun(fn)
}

func (c *continuation) ThenAsync(fn Runnable) Continuation {
	h := onSuccessContinuationHandlerAsync{fn: fn}
	return c.task.WhenComplete(h.invoke).AsContinuation()
}

func (c *continuation) CatchAsync(fn ErrorHandler) Continuation {
	return c.task.CatchAsync(fn).AsContinuation()
}

func (c *continuation) WhenCompleteAsync(fn func(error)) Continuation {
	r := &continuationRequestAdapterAsync{fn: fn}
	return c.task.WhenComplete(r.invoke).AsContinuation()
}

func (c *continuation) OnCompleteAsync(fn Handler[x.Void]) {
	c.task.OnCompleteAsync(fn)
}

func (c *continuation) OnCompleteRunAsync(fn Runnable) {
	c.task.OnCompleteRunAsync(fn)
}

func (c *continuation) AsContinuation() Continuation {
	return c.task.AsContinuation()
}

func (c *continuation) AsAsync() Continuation {
	return c.task.AsContinuationAsync()
}

func (c *continuation) IsAsync() bool {
	return c.task.IsAsync()
}
