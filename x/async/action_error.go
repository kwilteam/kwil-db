package async

import (
	"context"
	. "kwil/x"
	"kwil/x/errx"
)

type action_err struct {
	err error
}

func (_ *action_err) Fail(_ error) bool                           { return false }
func (_ *action_err) Complete() bool                              { return false }
func (_ *action_err) CompleteOrFail(_ error) bool                 { return false }
func (_ *action_err) Cancel() bool                                { return false }
func (_ *action_err) IsDone() bool                                { return true }
func (_ *action_err) IsError() bool                               { return true }
func (a *action_err) IsCancelled() bool                           { return errx.IsCancelled(a.err) }
func (_ *action_err) IsErrorOrCancelled() bool                    { return true }
func (_ *action_err) Await(_ context.Context) bool                { return true }
func (a *action_err) GetError() error                             { return a.err }
func (_ *action_err) DoneCh() <-chan Void                         { return ClosedChanVoid() }
func (a *action_err) Then(_ func()) Action                        { return a }
func (a *action_err) ThenCh(_ chan Void) Action                   { return a }
func (a *action_err) Catch(fn func(error)) Action                 { fn(a.err); return a }
func (a *action_err) CatchCh(ch chan error) Action                { ch <- a.err; return a }
func (a *action_err) WhenComplete(fn func(error)) Action          { fn(a.err); return a }
func (a *action_err) WhenCompleteCh(ch chan *Result[Void]) Action { return a._whenCompleteCh(ch) }
func (a *action_err) OnComplete(fn *Continuation[Void])           { fn.invoke(Void{}, a.err) }
func (a *action_err) AsAction() Action                            { return a }
func (a *action_err) AsListenable() Listenable[Void]              { return a }
func (a *action_err) AsAsync(e Executor) Action                   { return a._asAsync(e) }
func (_ *action_err) IsAsync() bool                               { return false }
func (a *action_err) ThenCatchFinally(fn *ContinuationA) Action {
	fn.invoke(a.err)
	return a
}

func (a *action_err) OnCompleteA(c *ContinuationA) {
	c.invoke(a.err)
}

func (a *action_err) _asAsync(e Executor) Action {
	if e != nil {
		e = DefaultExecutor()
	}

	c := _newAction()
	e.Execute(func() {
		a.Fail(a.err)
	})

	return c
}

func (a *action_err) _whenCompleteCh(ch chan *Result[Void]) Action {
	ch <- ResultError[Void](a.err)
	return a
}
