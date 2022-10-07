package rx

import (
	"context"
	"kwil/x"
)

type task_value[T any] struct {
	value T
}

func (_ *task_value[T]) GetError() error                             { return nil }
func (r *task_value[T]) Get() T                                      { return r.value }
func (_ *task_value[T]) IsError() bool                               { return false }
func (_ *task_value[T]) IsCancelled() bool                           { return false }
func (_ *task_value[T]) IsErrorOrCancelled() bool                    { return false }
func (_ *task_value[T]) IsDone() bool                                { return true }
func (_ *task_value[T]) DoneChan() <-chan x.Void                     { return x.ClosedChan }
func (_ *task_value[T]) Fail(error) bool                             { return false }
func (_ *task_value[T]) Complete(T) bool                             { return false }
func (_ *task_value[T]) CompleteOrFail(_ T, _ error) bool            { return false }
func (_ *task_value[T]) Cancel() bool                                { return false }
func (r *task_value[T]) GetOrError() (T, error)                      { return r.value, nil }
func (r *task_value[T]) Await(_ context.Context) (ok bool)           { return true }
func (r *task_value[T]) Then(fn ValueHandler[T]) Task[T]             { fn(r.value); return r }
func (r *task_value[T]) Catch(ErrorHandler) Task[T]                  { return r }
func (r *task_value[T]) Handle(fn func(T, error) (T, error)) Task[T] { return r._handle(fn) }
func (r *task_value[T]) Compose(fn func(T, error) Task[T]) Task[T]   { return fn(r.value, nil) }
func (r *task_value[T]) WhenComplete(fn Handler[T]) Task[T]          { fn(r.value, nil); return r }
func (r *task_value[T]) OnComplete(fn *Completion[T])                { fn.Invoke(r.value, nil) }
func (r *task_value[T]) AsContinuation() Continuation                { return &cont_value{} }
func (r *task_value[T]) AsContinuationAsync() Continuation           { return r._asContinuationAsync() }
func (r *task_value[T]) AsAsync() Task[T]                            { return r._asAsync() }
func (r *task_value[T]) IsAsync() bool                               { return false }
func (r *task_value[T]) ThenCatchFinally(fn *Completion[T]) Task[T] {
	return r.WhenComplete(fn.Invoke)
}

func (r *task_value[T]) _asContinuationAsync() Continuation {
	c := NewContinuationAsync()
	c.Complete()
	return c
}

func (r *task_value[T]) _asAsync() Task[T] {
	t := NewTaskAsync[T]()
	t.Complete(r.value)
	return t
}

func (r *task_value[T]) _handle(fn func(T, error) (T, error)) Task[T] {
	v, e := fn(r.value, nil)
	if e != nil {
		return Failure[T](e)
	} else {
		return Success(v)
	}
}
