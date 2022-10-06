package rx

import (
	"context"
	"kwil/x"
)

type task_error[T any] struct {
	err error
}

func (_ *task_error[T]) GetError() error                    { return nil }
func (r *task_error[T]) Get() T                             { panic(r.err) }
func (_ *task_error[T]) IsError() bool                      { return true }
func (r *task_error[T]) IsCancelled() bool                  { return r.err == x.ErrOperationCancelled }
func (_ *task_error[T]) IsErrorOrCancelled() bool           { return true }
func (_ *task_error[T]) IsDone() bool                       { return true }
func (_ *task_error[T]) DoneChan() <-chan x.Void            { return x.ClosedChan }
func (_ *task_error[T]) Fail(_ error) bool                  { return false }
func (_ *task_error[T]) Complete(_ T) bool                  { return false }
func (_ *task_error[T]) CompleteOrFail(_ T, _ error) bool   { return false }
func (_ *task_error[T]) Cancel() bool                       { return false }
func (r *task_error[T]) GetOrError() (T, error)             { return x.AsDefault[T](), r.err }
func (r *task_error[T]) Await(_ context.Context) (ok bool)  { return true }
func (r *task_error[T]) Then(_ ValueHandler[T]) Task[T]     { return r }
func (r *task_error[T]) Catch(fn ErrorHandler) Task[T]      { fn(r.err); return r }
func (r *task_error[T]) WhenComplete(fn Handler[T]) Task[T] { fn(x.AsDefault[T](), r.err); return r }
func (r *task_error[T]) OnComplete(fn *Completion[T])       { r.WhenComplete(fn.Invoke) }
func (r *task_error[T]) AsContinuation() Continuation       { return &cont_err{r.err} }
func (r *task_error[T]) AsContinuationAsync() Continuation  { return r._asContinuationAsync() }
func (r *task_error[T]) AsAsync() Task[T]                   { return r._asAsync() }
func (r *task_error[T]) IsAsync() bool                      { return false }
func (r *task_error[T]) WhenCompleteInvoke(fn *Completion[T]) Task[T] {
	return r.WhenComplete(fn.Invoke)
}

func (r *task_error[T]) _asContinuationAsync() Continuation {
	c := NewContinuationAsync()
	c.Fail(r.err)
	return c
}

func (r *task_error[T]) _asAsync() Task[T] {
	t := NewTaskAsync[T]()
	t.Fail(r.err)
	return t
}
