package async

import (
	"context"
	. "kwil/x"
	"kwil/x/errx"
)

type task_error[T any] struct {
	err error
}

func (_ *task_error[T]) GetError() error                             { return nil }
func (r *task_error[T]) Get() T                                      { panic(r.err) }
func (_ *task_error[T]) IsError() bool                               { return true }
func (r *task_error[T]) IsCancelled() bool                           { return errx.IsCancelled(r.err) }
func (_ *task_error[T]) IsErrorOrCancelled() bool                    { return true }
func (_ *task_error[T]) IsDone() bool                                { return true }
func (_ *task_error[T]) DoneCh() <-chan Void                         { return ClosedChanVoid() }
func (_ *task_error[T]) Fail(_ error) bool                           { return false }
func (_ *task_error[T]) Complete(_ T) bool                           { return false }
func (_ *task_error[T]) CompleteOrFail(_ T, _ error) bool            { return false }
func (_ *task_error[T]) Cancel() bool                                { return false }
func (r *task_error[T]) GetOrError() (T, error)                      { return AsDefault[T](), r.err }
func (r *task_error[T]) Await(_ context.Context) (ok bool)           { return true }
func (r *task_error[T]) Then(_ func(T)) Task[T]                      { return r }
func (r *task_error[T]) ThenCh(_ chan T) Task[T]                     { return r }
func (r *task_error[T]) Catch(fn func(error)) Task[T]                { fn(r.err); return r }
func (r *task_error[T]) CatchCh(ch chan error) Task[T]               { ch <- r.err; return r }
func (r *task_error[T]) Handle(fn func(T, error) (T, error)) Task[T] { return r._handle(fn) }
func (r *task_error[T]) Compose(fn func(T, error) Task[T]) Task[T]   { return r._compose(fn) }
func (r *task_error[T]) ComposeA(fn func(T, error) Action) Action    { return fn(AsDefault[T](), r.err) }
func (r *task_error[T]) WhenComplete(fn func(T, error)) Task[T]      { fn(AsDefault[T](), r.err); return r }
func (r *task_error[T]) OnComplete(fn *Continuation[T])              { r.WhenComplete(fn.invoke) }
func (r *task_error[T]) AsAction() Action                            { return &action_err{r.err} }
func (r *task_error[T]) AsListenable() Listenable[T]                 { return r }
func (r *task_error[T]) AsAsync(e Executor) Task[T]                  { return r._asAsync(e) }
func (r *task_error[T]) IsAsync() bool                               { return false }
func (r *task_error[T]) ThenCatchFinally(fn *Continuation[T]) Task[T] {
	return r.WhenComplete(fn.invoke)
}
func (r *task_error[T]) WhenCompleteCh(ch chan *Result[T]) Task[T] {
	ch <- ResultError[T](r.err)
	return r
}

func (r *task_error[T]) _asAsync(e Executor) Task[T] {
	if e != nil {
		e = DefaultExecutor()
	}

	t := newTask[T]()

	e.Execute(func() { t.Fail(r.err) })

	return t
}

func (r *task_error[T]) _handle(fn func(T, error) (T, error)) Task[T] {
	v, e := fn(AsDefault[T](), r.err)
	if e != nil {
		return FailedTask[T](e)
	} else {
		return CompletedTask(v)
	}
}

func (r *task_error[T]) _compose(fn func(T, error) Task[T]) Task[T] {
	return fn(AsDefault[T](), r.err)
}
