package rx

import (
	"context"
	"kwil/x"
)

type cont_err struct {
	err error
}

func (_ *cont_err) Fail(_ error) bool                        { return false }
func (_ *cont_err) Complete() bool                           { return false }
func (_ *cont_err) CompleteOrFail(_ error) bool              { return false }
func (_ *cont_err) Cancel() bool                             { return false }
func (_ *cont_err) IsDone() bool                             { return true }
func (_ *cont_err) IsError() bool                            { return true }
func (c *cont_err) IsCancelled() bool                        { return c.err == x.ErrOperationCancelled }
func (_ *cont_err) IsErrorOrCancelled() bool                 { return true }
func (_ *cont_err) Await(_ context.Context) bool             { return true }
func (c *cont_err) GetError() error                          { return c.err }
func (_ *cont_err) DoneChan() <-chan x.Void                  { return x.ClosedChan }
func (c *cont_err) Then(_ Runnable) Continuation             { return c }
func (c *cont_err) Catch(fn ErrorHandler) Continuation       { fn(c.err); return c }
func (c *cont_err) WhenComplete(fn func(error)) Continuation { fn(c.err); return c }
func (c *cont_err) OnComplete(fn *Completion[x.Void])        { fn.Invoke(x.Void{}, c.err) }
func (c *cont_err) AsContinuation() Continuation             { return c }
func (c *cont_err) AsAsync() Continuation                    { return c._asAsync() }
func (c *cont_err) IsAsync() bool                            { return false }
func (c *cont_err) WhenCompleteInvoke(fn *CompletionC) Continuation {
	fn.Invoke(c.err)
	return c
}

func (c *cont_err) _catchAsync(fn ErrorHandler) Continuation {
	go func(fn ErrorHandler, err error) { fn(err) }(fn, c.err)
	return c
}

func (c *cont_err) _whenCompleteAsync(fn func(error)) Continuation {
	go func(fn ErrorHandler, err error) { fn(err) }(fn, c.err)
	return c
}

func (c *cont_err) _whenCompleteRunAsync(fn Runnable) Continuation {
	go fn()
	return c
}

func (_ *cont_err) _onCompleteAsync(fn Handler[x.Void]) {
	go fn(x.Void{}, nil)
}

func (_ *cont_err) _onCompleteRunAsync(fn Runnable) {
	go fn()
}

func (c *cont_err) _asAsync() Continuation {
	t := NewContinuationAsync()
	t.Fail(c.err)
	return t
}
