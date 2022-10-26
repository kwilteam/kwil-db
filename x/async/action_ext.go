package async

import (
	"context"
	. "kwil/x"
)

type action struct {
	task *_task[Void]
}

func (a *action) Fail(err error) bool                         { return a.task.Fail(err) }
func (a *action) Complete() bool                              { return a.task.Complete(Void{}) }
func (a *action) CompleteOrFail(err error) bool               { return a.task.CompleteOrFail(Void{}, err) }
func (a *action) Cancel() bool                                { return a.task.Cancel() }
func (a *action) IsDone() bool                                { return a.task.IsDone() }
func (a *action) IsError() bool                               { return a.task.IsError() }
func (a *action) IsCancelled() bool                           { return a.task.IsCancelled() }
func (a *action) IsErrorOrCancelled() bool                    { return a.task.IsErrorOrCancelled() }
func (a *action) Await(ctx context.Context) bool              { return a.task.Await(ctx) }
func (a *action) GetError() error                             { return a.task.GetError() }
func (a *action) DoneCh() <-chan Void                         { return a.task.DoneCh() }
func (a *action) Then(fn func()) Action                       { return a._then(fn) }
func (a *action) ThenCh(ch chan Void) Action                  { return a.task.ThenCh(ch).AsAction() }
func (a *action) Catch(fn func(error)) Action                 { return a.task.Catch(fn).AsAction() }
func (a *action) CatchCh(ch chan error) Action                { return a.task.CatchCh(ch).AsAction() }
func (a *action) ThenCatchFinally(fn *ContinuationA) Action   { return a._whenComplete(fn.invoke) }
func (a *action) WhenComplete(fn func(error)) Action          { return a._whenComplete(fn) }
func (a *action) WhenCompleteCh(ch chan *Result[Void]) Action { return a._whenCompleteCh(ch) }
func (a *action) OnComplete(fn *Continuation[Void])           { a.task.WhenComplete(fn.invoke).AsAction() }
func (a *action) AsAction() Action                            { return a.task.AsAction() }
func (a *action) AsListenable() Listenable[Void]              { return a.task.AsListenable() }
func (a *action) AsAsync(e Executor) Action                   { return a._asAsync(e) }
func (a *action) IsAsync() bool                               { return a.task.IsAsync() }
func (a *action) OnCompleteA(c *ContinuationA) {
	a._whenComplete(c.invoke)
}

func (a *action) _then(fn func()) Action {
	h := onSuccessContinuationHandler{fn: fn}
	return a.task.WhenComplete(h.invoke).AsAction()
}

func (a *action) _whenComplete(fn func(error)) Action {
	r := &continuationRequestAdapter{fn: fn}
	return a.task.WhenComplete(r.invoke).AsAction()
}

func (a *action) _whenCompleteCh(ch chan *Result[Void]) Action {
	return a.task.WhenCompleteCh(ch).AsAction()
}

func (a *action) _asAsync(e Executor) Action {
	if e == nil {
		e = AsyncExecutor()
	}

	n := _newAction()
	h := executorHandler[Void]{task: n.task, e: e}
	a.task._addHandlerNoReturn(h.invoke)
	return n
}

func _newAction() *action      { return &action{newTask[Void]()} }
func _newActionAsync() *action { return &action{newTaskAsync[Void]()} }
