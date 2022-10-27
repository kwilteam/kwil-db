package async

import (
	"context"
	. "kwil/x"
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
func (_ *task_value[T]) DoneCh() <-chan Void                         { return ClosedChanVoid() }
func (_ *task_value[T]) Fail(error) bool                             { return false }
func (_ *task_value[T]) Complete(T) bool                             { return false }
func (_ *task_value[T]) CompleteOrFail(_ T, _ error) bool            { return false }
func (_ *task_value[T]) Cancel() bool                                { return false }
func (r *task_value[T]) GetOrError() (T, error)                      { return r.value, nil }
func (r *task_value[T]) Await(_ context.Context) (ok bool)           { return true }
func (r *task_value[T]) Then(fn func(T)) Task[T]                     { fn(r.value); return r }
func (r *task_value[T]) ThenCh(ch chan T) Task[T]                    { ch <- r.value; return r }
func (r *task_value[T]) CatchCh(_ chan error) Task[T]                { return r }
func (r *task_value[T]) Catch(func(error)) Task[T]                   { return r }
func (r *task_value[T]) Handle(fn func(T, error) (T, error)) Task[T] { return r._handle(fn) }
func (r *task_value[T]) Compose(fn func(T, error) Task[T]) Task[T]   { return fn(r.value, nil) }
func (r *task_value[T]) WhenComplete(fn func(T, error)) Task[T]      { fn(r.value, nil); return r }
func (r *task_value[T]) OnComplete(fn *Continuation[T])              { fn.invoke(r.value, nil) }
func (r *task_value[T]) AsAction() Action                            { return &action_value{} }
func (r *task_value[T]) AsListenable() Listenable[T]                 { return r }
func (r *task_value[T]) AsAsync(e Executor) Task[T]                  { return r._asAsync(e) }
func (r *task_value[T]) IsAsync() bool                               { return false }
func (r *task_value[T]) ThenCatchFinally(fn *Continuation[T]) Task[T] {
	return r.WhenComplete(fn.invoke)
}
func (r *task_value[T]) WhenCompleteCh(ch chan *Result[T]) Task[T] {
	ch <- ResultSuccess(r.value)
	return r
}

func (r *task_value[T]) _asAsync(e Executor) Task[T] {
	if e != nil {
		e = DefaultExecutor()
	}

	t := newTask[T]()

	e.Execute(func() { t.Complete(r.value) })

	return t
}

func (r *task_value[T]) _handle(fn func(T, error) (T, error)) Task[T] {
	v, e := fn(r.value, nil)
	if e != nil {
		return FailedTask[T](e)
	} else {
		return CompletedTask(v)
	}
}
