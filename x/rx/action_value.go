package rx

import (
	"context"
	. "kwil/x"
)

type action_value struct{}

func (_ *action_value) Fail(_ error) bool                  { return false }
func (_ *action_value) Complete() bool                     { return false }
func (_ *action_value) CompleteOrFail(_ error) bool        { return false }
func (_ *action_value) Cancel() bool                       { return false }
func (_ *action_value) IsDone() bool                       { return true }
func (_ *action_value) IsError() bool                      { return false }
func (_ *action_value) IsCancelled() bool                  { return false }
func (_ *action_value) IsErrorOrCancelled() bool           { return false }
func (_ *action_value) Await(_ context.Context) bool       { return true }
func (_ *action_value) GetError() error                    { return nil }
func (_ *action_value) DoneChan() <-chan Void              { return ClosedChanVoid() }
func (c *action_value) Then(fn func()) Action              { fn(); return c }
func (c *action_value) Catch(_ func(error)) Action         { return c }
func (c *action_value) OnComplete(fn *ContinuationT[Void]) { fn.invoke(Void{}, nil) }
func (c *action_value) WhenComplete(fn func(error)) Action { fn(nil); return c }
func (c *action_value) AsAction() Action                   { return c }
func (c *action_value) AsListenable() Listenable[Void]     { return c }
func (c *action_value) AsAsync(e Executor) Action          { return c._asAsync(e) }
func (c *action_value) IsAsync() bool                      { return false }
func (c *action_value) ThenCatchFinally(fn *ContinuationA) Action {
	fn.invoke(nil)
	return c
}

func (c *action_value) _asAsync(e Executor) Action {
	if e != nil {
		e = asyncExecutor
	}

	a := _newAction()

	e.Execute(func() { a.Complete() })

	return a
}
