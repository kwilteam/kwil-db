package rx

import (
	"context"
	"kwil/x/utils"
	"sync/atomic"
)

func (r *Task[T]) _cancel() bool {
	return r.completeOrFail(utils.AsDefault[T](), ErrCancelled)
}

func (r *Task[T]) _completeOrFail(value T, err error) bool {
	return r.completeOrFail(value, err)
}

func (r *Task[T]) _complete(val T) bool {
	return r.completeOrFail(val, nil)
}

func (r *Task[T]) _fail(err error) bool {
	return r.completeOrFail(utils.AsDefault[T](), err)
}

func (r *Task[T]) _isDone() bool {
	current := atomic.LoadUint32(r.status)
	return isDone(current)
}

func (r *Task[T]) _addHandlerNoReturn(fn Handler[T]) {
	r.addFnHandler(fn, true)
}

func (r *Task[T]) _addHandler(fn Handler[T]) *Task[T] {
	return r.addFnHandler(fn, false)
}

func (r *Task[T]) _await(ctx context.Context) (ok bool) {
	if isCompletedOrigin(*r.status) {
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

func (r *Task[T]) _isError() bool {
	current := atomic.LoadUint32(r.status)
	return hasError(current)
}

func (r *Task[T]) _isCancelled() bool {
	current := atomic.LoadUint32(r.status)
	return isCancelled(current)
}

func (r *Task[T]) _isErrorOrCancelled() bool {
	return r._isError()
}

func (r *Task[T]) _getOrError() (T, error) {
	<-r._doneChan()

	current := atomic.LoadUint32(r.status)
	return r.loadValueOrErrorUnsafe(current)
}

func (r *Task[T]) _doneChan() <-chan struct{} {
	if isCompletedOrigin(*r.status) {
		return _closedChan
	}

	return r.getOrAddDoneBlockChanHandler()
}

func (r *Task[T]) _asContinuation(async bool) *Continuation {
	status := utils.IfElse(async, _ASYNC_CONTINUATIONS, uint32(0))
	h := asContinuationHandler[T]{&Continuation{task: &Task[struct{}]{status: &status}}}
	r._onComplete(h.invoke)
	return h.c
}

func (r *Task[T]) _async() *Task[T] {
	task := NewTaskAsync[T]()
	r._addHandlerNoReturn(task.setEitherNoReturn)
	return task
}

func (r *Task[T]) _then(fn ValueHandler[T]) *Task[T] {
	h := &onSuccessHandler[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *Task[T]) _catch(fn ErrorHandler) *Task[T] {
	h := &onErrorHandler[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *Task[T]) _whenComplete(fn Handler[T]) *Task[T] {
	return r._addHandler(fn)
}

func (r *Task[T]) _onComplete(fn Handler[T]) {
	r._addHandlerNoReturn(fn)
}

func (r *Task[T]) _onCompleteRun(fn Runnable) {
	h := onCompleteRunHandler[T]{fn: fn}
	r._addHandlerNoReturn(h.invoke)
}

func (r *Task[T]) _thenAsync(fn ValueHandler[T]) *Task[T] {
	h := &onSuccessHandlerAsync[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *Task[T]) _catchErrorAsync(fn ErrorHandler) *Task[T] {
	h := &onErrorHandlerAsync[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *Task[T]) _whenCompleteAsync(fn Handler[T]) *Task[T] {
	h := &onCompleteHandlerAsync[T]{fn: fn}
	return r._addHandler(h.invoke)
}

func (r *Task[T]) _onCompleteAsync(fn Handler[T]) {
	h := &onCompleteHandlerAsync[T]{fn: fn}
	r._addHandlerNoReturn(h.invoke)
}

func (r *Task[T]) _onCompleteRunAsync(fn Runnable) {
	h := onCompleteRunHandlerAsync[T]{fn: fn}
	r._addHandlerNoReturn(h.invoke)
}

func (r *Task[T]) _getError() error {
	_, err := r._getOrError()
	return err
}

func (r *Task[T]) _get() T {
	v, e := r._getOrError()
	if e != nil {
		panic(e)
	}

	return v
}
