package rx

import (
	"context"
	. "kwil/x"
	"kwil/x/errx"
	"sync/atomic"
	"unsafe"
)

func (r *_task[T]) _cancel() bool {
	return r.completeOrFail(AsDefault[T](), errx.ErrOperationCancelled())
}

func (r *_task[T]) _completeOrFailNoReturn(value T, err error) {
	r.completeOrFail(value, err)
}

func (r *_task[T]) _completeOrFail(value T, err error) bool {
	return r.completeOrFail(value, err)
}

func (r *_task[T]) _complete(val T) bool {
	return r.completeOrFail(val, nil)
}

func (r *_task[T]) _fail(err error) bool {
	return r.completeOrFail(AsDefault[T](), err)
}

func (r *_task[T]) _isDone() bool {
	current := atomic.LoadUint32(&r.status)
	return isDone(current)
}

func (r *_task[T]) _addHandlerNoReturn(fn func(T, error)) {
	r.addFnHandler(fn, true)
}

func (r *_task[T]) _addHandler(fn func(T, error)) *_task[T] {
	return r.addFnHandler(fn, false)
}

func (r *_task[T]) _await(ctx context.Context) (ok bool) {
	if ctx == nil {
		<-r._doneChan()
		return true
	}

	select {
	case <-ctx.Done():
		return false
	case <-r._doneChan():
		return true
	}
}

func (r *_task[T]) _isError() bool {
	current := atomic.LoadUint32(&r.status)
	return hasError(current)
}

func (r *_task[T]) _isCancelled() bool {
	current := atomic.LoadUint32(&r.status)
	return isCancelled(current)
}

func (r *_task[T]) _isErrorOrCancelled() bool {
	return r._isError()
}

func (r *_task[T]) _getOrError() (T, error) {
	<-r._doneChan()

	current := atomic.LoadUint32(&r.status)
	return r.loadValueOrError(current)
}

func (r *_task[T]) _doneChan() <-chan Void {
	return r.getOrAddDoneChan()
}

func (r *_task[T]) _asAction() Action {
	a := asActionHandler[T]{_newAction()}
	r._addHandlerNoReturn(a.invoke)
	return a.action
}

func (r *_task[T]) _async(e Executor) Task[T] {
	if e == nil {
		e = asyncExecutor
	}

	h := executorHandler[T]{task: newTask[T](), e: e}
	r._addHandlerNoReturn(h.invoke)

	return h.task
}

func (r *_task[T]) _then(fn func(T)) Task[T] {
	h := &onSuccessHandler[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *_task[T]) _thenCh(ch chan T) Task[T] {
	return r._addHandler(func(v T, _ error) {
		ch <- v
	})
}

func (r *_task[T]) _catch(fn func(error)) Task[T] {
	h := &onErrorHandler[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *_task[T]) _catchCh(ch chan error) Task[T] {
	return r._addHandler(func(_ T, err error) {
		ch <- err
	})
}

func (r *_task[T]) _handle(fn func(T, error) (T, error)) Task[T] {
	h := onHandle[T]{newTask[T](), fn}
	r._addHandlerNoReturn(h.invoke)
	return h.task
}

func (r *_task[T]) _compose(fn func(T, error) Task[T]) Task[T] {
	h := onCompose[T]{newTask[T](), fn}
	r._addHandlerNoReturn(h.invoke)
	return h.task
}

func (r *_task[T]) _whenComplete(fn func(T, error)) Task[T] {
	return r._addHandler(fn)
}

func (r *_task[T]) _whenCompleteCh(ch chan *Result[T]) Task[T] {
	return r._addHandler(func(v T, err error) {
		if err != nil {
			ch <- ResultSuccess[T](v)
		} else {
			ch <- ResultError[T](err)
		}
	})
}

func (r *_task[T]) _getError() error {
	_, err := r._getOrError()
	return err
}

func (r *_task[T]) _get() T {
	v, e := r._getOrError()
	if e != nil {
		panic(e)
	}

	return v
}

func newTask[T any]() *_task[T] {
	var state unsafe.Pointer
	return &_task[T]{status: uint32(0), state: state}
}

func newTaskAsync[T any]() *_task[T] {
	var state unsafe.Pointer
	return &_task[T]{
		status: _ASYNC_CONTINUATIONS,
		state:  state,
	}
}
