package rx

import (
	"context"
	"kwil/x"
	"sync/atomic"
	"unsafe"
)

func (r *task[T]) _cancel() bool {
	return r.completeOrFail(x.AsDefault[T](), x.ErrOperationCancelled)
}

func (r *task[T]) _completeOrFail(value T, err error) bool {
	return r.completeOrFail(value, err)
}

func (r *task[T]) _complete(val T) bool {
	return r.completeOrFail(val, nil)
}

func (r *task[T]) _fail(err error) bool {
	return r.completeOrFail(x.AsDefault[T](), err)
}

func (r *task[T]) _isDone() bool {
	current := atomic.LoadUint32(&r.status)
	return isDone(current)
}

func (r *task[T]) _addHandlerNoReturn(fn Handler[T]) {
	r.addFnHandler(fn, true)
}

func (r *task[T]) _addHandler(fn Handler[T]) *task[T] {
	return r.addFnHandler(fn, false)
}

func (r *task[T]) _await(ctx context.Context) (ok bool) {
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

func (r *task[T]) _isError() bool {
	current := atomic.LoadUint32(&r.status)
	return hasError(current)
}

func (r *task[T]) _isCancelled() bool {
	current := atomic.LoadUint32(&r.status)
	return isCancelled(current)
}

func (r *task[T]) _isErrorOrCancelled() bool {
	return r._isError()
}

func (r *task[T]) _getOrError() (T, error) {
	<-r._doneChan()

	current := atomic.LoadUint32(&r.status)
	return r.loadValueOrErrorUnsafe(current)
}

func (r *task[T]) _doneChan() <-chan x.Void {
	return r.getOrAddDoneChan()
}

func (r *task[T]) _asContinuation(async bool) Continuation {
	var c Continuation
	if async {
		c = NewContinuationAsync()
	} else {
		c = NewContinuation()
	}

	r.WhenComplete(func(_ T, err error) {
		c.CompleteOrFail(err)
	})

	return c
}

func (r *task[T]) _async() *task[T] {
	task := newTaskAsync[T]()
	r._addHandlerNoReturn(task.setEitherNoReturn)
	return task
}

func (r *task[T]) _then(fn ValueHandler[T]) *task[T] {
	h := &onSuccessHandler[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *task[T]) _catch(fn ErrorHandler) *task[T] {
	h := &onErrorHandler[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *task[T]) _whenComplete(fn Handler[T]) *task[T] {
	return r._addHandler(fn)
}

func (r *task[T]) _getError() error {
	_, err := r._getOrError()
	return err
}

func (r *task[T]) _get() T {
	v, e := r._getOrError()
	if e != nil {
		panic(e)
	}

	return v
}

func newTask[T any]() *task[T] {
	var state unsafe.Pointer
	return &task[T]{status: uint32(0), state: state}
}

func newTaskAsync[T any]() *task[T] {
	var state unsafe.Pointer
	return &task[T]{
		status: _ASYNC_CONTINUATIONS,
		state:  state,
	}
}
