package rx

import (
	"context"
	"kwil/x"
)

type cont_value struct{}

func (_ *cont_value) Fail(_ error) bool                        { return false }
func (_ *cont_value) Complete() bool                           { return false }
func (_ *cont_value) CompleteOrFail(_ error) bool              { return false }
func (_ *cont_value) Cancel() bool                             { return false }
func (_ *cont_value) IsDone() bool                             { return true }
func (_ *cont_value) IsError() bool                            { return false }
func (_ *cont_value) IsCancelled() bool                        { return false }
func (_ *cont_value) IsErrorOrCancelled() bool                 { return false }
func (_ *cont_value) Await(_ context.Context) bool             { return true }
func (_ *cont_value) GetError() error                          { return nil }
func (_ *cont_value) DoneChan() <-chan x.Void                  { return x.ClosedChan }
func (c *cont_value) Then(fn Runnable) Continuation            { fn(); return c }
func (c *cont_value) Catch(_ ErrorHandler) Continuation        { return c }
func (c *cont_value) OnComplete(fn *Completion[x.Void])        { fn.Invoke(x.Void{}, nil) }
func (c *cont_value) WhenComplete(fn func(error)) Continuation { fn(nil); return c }
func (c *cont_value) AsContinuation() Continuation             { return c }
func (c *cont_value) AsAsync() Continuation                    { return c.AsAsync() }
func (c *cont_value) IsAsync() bool                            { return false }
func (c *cont_value) ThenCatchFinally(fn *CompletionC) Continuation {
	fn.Invoke(nil)
	return c
}

func (c *cont_value) _asAsync() Continuation {
	t := NewContinuationAsync()
	t.Complete()
	return t
}
