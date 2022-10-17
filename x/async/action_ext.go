package async

import (
	"context"
	. "kwil/x"
)

type action struct {
	task *_task[Void]
}

func (c *action) Fail(err error) bool                         { return c.task.Fail(err) }
func (c *action) Complete() bool                              { return c.task.Complete(Void{}) }
func (c *action) CompleteOrFail(err error) bool               { return c.task.CompleteOrFail(Void{}, err) }
func (c *action) Cancel() bool                                { return c.task.Cancel() }
func (c *action) IsDone() bool                                { return c.task.IsDone() }
func (c *action) IsError() bool                               { return c.task.IsError() }
func (c *action) IsCancelled() bool                           { return c.task.IsCancelled() }
func (c *action) IsErrorOrCancelled() bool                    { return c.task.IsErrorOrCancelled() }
func (c *action) Await(ctx context.Context) bool              { return c.task.Await(ctx) }
func (c *action) GetError() error                             { return c.task.GetError() }
func (c *action) DoneCh() <-chan Void                         { return c.task.DoneCh() }
func (c *action) Then(fn func()) Action                       { return c._then(fn) }
func (c *action) ThenCh(ch chan Void) Action                  { return c.task.ThenCh(ch).AsAction() }
func (c *action) Catch(fn func(error)) Action                 { return c.task.Catch(fn).AsAction() }
func (c *action) CatchCh(ch chan error) Action                { return c.task.CatchCh(ch).AsAction() }
func (c *action) ThenCatchFinally(fn *ContinuationA) Action   { return c._whenComplete(fn.invoke) }
func (c *action) WhenComplete(fn func(error)) Action          { return c._whenComplete(fn) }
func (c *action) WhenCompleteCh(ch chan *Result[Void]) Action { return c._whenCompleteCh(ch) }
func (c *action) OnComplete(fn *Continuation[Void])           { c.task.WhenComplete(fn.invoke).AsAction() }
func (c *action) AsAction() Action                            { return c.task.AsAction() }
func (c *action) AsListenable() Listenable[Void]              { return c.task.AsListenable() }
func (c *action) AsAsync(e Executor) Action                   { return c._asAsync(e) }
func (c *action) IsAsync() bool                               { return c.task.IsAsync() }

func (c *action) _then(fn func()) Action {
	h := onSuccessContinuationHandler{fn: fn}
	return c.task.WhenComplete(h.invoke).AsAction()
}

func (c *action) _whenComplete(fn func(error)) Action {
	r := &continuationRequestAdapter{fn: fn}
	return c.task.WhenComplete(r.invoke).AsAction()
}

func (c *action) _whenCompleteCh(ch chan *Result[Void]) Action {
	return c.task.WhenCompleteCh(ch).AsAction()
}

func (c *action) _asAsync(e Executor) Action {
	if e == nil {
		e = AsyncExecutor()
	}

	a := _newAction()
	h := executorHandler[Void]{task: a.task, e: e}
	c.task._addHandlerNoReturn(h.invoke)
	return a
}

func _newAction() *action      { return &action{newTask[Void]()} }
func _newActionAsync() *action { return &action{newTaskAsync[Void]()} }
