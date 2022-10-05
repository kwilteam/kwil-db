package rx

import (
	"context"
	"kwil/x/utils"
	"sync/atomic"
	"unsafe"
)

func (r *task[T]) _cancel() bool {
	return r.completeOrFail(utils.AsDefault[T](), ErrCancelled)
}

func (r *task[T]) _completeOrFail(value T, err error) bool {
	return r.completeOrFail(value, err)
}

func (r *task[T]) _complete(val T) bool {
	return r.completeOrFail(val, nil)
}

func (r *task[T]) _fail(err error) bool {
	return r.completeOrFail(utils.AsDefault[T](), err)
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
	if isCompletedOrigin(r.status) {
		return true
	}

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

func (r *task[T]) _doneChan() <-chan Void {
	if isCompletedOrigin(r.status) {
		return _closedChan
	}

	ch := make(chan Void)
	r._addHandlerNoReturn(func(_ T, _ error) {
		close(ch)
	})

	return ch
}

func (r *task[T]) _asContinuation(async bool) *Continuation {
	var state unsafe.Pointer
	status := utils.IfElse(async, _ASYNC_CONTINUATIONS, uint32(0))
	h := asContinuationHandler[T]{&Continuation{task: &task[Void]{
		status: status,
		state:  state,
	}}}
	r._onComplete(h.invoke)
	return h.c
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

func (r *task[T]) _onComplete(fn Handler[T]) {
	r._addHandlerNoReturn(fn)
}

func (r *task[T]) _onCompleteRun(fn Runnable) {
	h := onCompleteRunHandler[T]{fn: fn}
	r._addHandlerNoReturn(h.invoke)
}

func (r *task[T]) _thenAsync(fn ValueHandler[T]) *task[T] {
	h := &onSuccessHandlerAsync[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *task[T]) _catchErrorAsync(fn ErrorHandler) *task[T] {
	h := &onErrorHandlerAsync[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *task[T]) _whenCompleteAsync(fn Handler[T]) *task[T] {
	h := &onCompleteHandlerAsync[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *task[T]) _onCompleteAsync(fn Handler[T]) {
	h := &onCompleteHandlerAsync[T]{fn: fn}
	r._addHandlerNoReturn(h.invoke)
}

func (r *task[T]) _onCompleteRunAsync(fn Runnable) {
	h := onCompleteRunHandlerAsync[T]{fn: fn}
	r._addHandlerNoReturn(h.invoke)
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
