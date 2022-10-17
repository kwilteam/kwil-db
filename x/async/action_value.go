package async

import (
	"context"
	. "kwil/x"
)

type action_value struct{}

func (_ *action_value) Fail(_ error) bool                           { return false }
func (_ *action_value) Complete() bool                              { return false }
func (_ *action_value) CompleteOrFail(_ error) bool                 { return false }
func (_ *action_value) Cancel() bool                                { return false }
func (_ *action_value) IsDone() bool                                { return true }
func (_ *action_value) IsError() bool                               { return false }
func (_ *action_value) IsCancelled() bool                           { return false }
func (_ *action_value) IsErrorOrCancelled() bool                    { return false }
func (_ *action_value) Await(_ context.Context) bool                { return true }
func (_ *action_value) GetError() error                             { return nil }
func (_ *action_value) DoneCh() <-chan Void                         { return ClosedChanVoid() }
func (a *action_value) Then(fn func()) Action                       { fn(); return a }
func (a *action_value) ThenCh(ch chan Void) Action                  { ch <- Void{}; return a }
func (a *action_value) Catch(_ func(error)) Action                  { return a }
func (a *action_value) CatchCh(_ chan error) Action                 { return a }
func (a *action_value) OnComplete(fn *Continuation[Void])           { fn.invoke(Void{}, nil) }
func (a *action_value) WhenComplete(fn func(error)) Action          { fn(nil); return a }
func (a *action_value) WhenCompleteCh(ch chan *Result[Void]) Action { return a._whenCompleteCh(ch) }
func (a *action_value) AsAction() Action                            { return a }
func (a *action_value) AsListenable() Listenable[Void]              { return a }
func (a *action_value) AsAsync(e Executor) Action                   { return a._asAsync(e) }
func (a *action_value) IsAsync() bool                               { return false }
func (a *action_value) ThenCatchFinally(fn *ContinuationA) Action {
	fn.invoke(nil)
	return a
}

func (a *action_value) _asAsync(e Executor) Action {
	if e != nil {
		e = AsyncExecutor()
	}

	n := _newAction()

	e.Execute(func() { a.Complete() })

	return n
}
func (a *action_value) _whenCompleteCh(ch chan *Result[Void]) Action {
	ch <- ResultSuccess(Void{})
	return a
}
