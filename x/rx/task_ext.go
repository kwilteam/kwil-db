package rx

import (
	"context"
	"kwil/x"
	"kwil/x/utils"
	"sync/atomic"
	"unsafe"
)

type task[T any] struct {
	status uint32
	state  unsafe.Pointer
}

func (r *task[T]) GetError() error                         { return r._getError() }
func (r *task[T]) Get() T                                  { return r._get() }
func (r *task[T]) IsError() bool                           { return r._isError() }
func (r *task[T]) IsCancelled() bool                       { return r._isCancelled() }
func (r *task[T]) IsErrorOrCancelled() bool                { return r._isErrorOrCancelled() }
func (r *task[T]) IsDone() bool                            { return r._isDone() }
func (r *task[T]) DoneChan() <-chan x.Void                 { return r._doneChan() }
func (r *task[T]) Fail(err error) bool                     { return r._fail(err) }
func (r *task[T]) Complete(value T) bool                   { return r._complete(value) }
func (r *task[T]) CompleteOrFail(value T, err error) bool  { return r._completeOrFail(value, err) }
func (r *task[T]) Cancel() bool                            { return r._cancel() }
func (r *task[T]) GetOrError() (T, error)                  { return r._getOrError() }
func (r *task[T]) Await(ctx context.Context) (ok bool)     { return r._await(ctx) }
func (r *task[T]) Then(fn ValueHandler[T]) Task[T]         { return r._then(fn) }
func (r *task[T]) Catch(fn ErrorHandler) Task[T]           { return r._catch(fn) }
func (r *task[T]) WhenComplete(fn Handler[T]) Task[T]      { return r._whenComplete(fn) }
func (r *task[T]) OnCompleteRun(fn Runnable)               { r._onCompleteRun(fn) }
func (r *task[T]) OnComplete(fn Handler[T])                { r._onComplete(fn) }
func (r *task[T]) ThenAsync(fn ValueHandler[T]) Task[T]    { return r._thenAsync(fn) }
func (r *task[T]) CatchAsync(fn ErrorHandler) Task[T]      { return r._catchErrorAsync(fn) }
func (r *task[T]) WhenCompleteAsync(fn Handler[T]) Task[T] { return r._whenCompleteAsync(fn) }
func (r *task[T]) OnCompleteAsync(fn Handler[T])           { r._onCompleteAsync(fn) }
func (r *task[T]) OnCompleteRunAsync(fn Runnable)          { r._onCompleteRunAsync(fn) }
func (r *task[T]) AsContinuation() Continuation            { return r._asContinuation(false) }
func (r *task[T]) AsContinuationAsync() Continuation       { return r._asContinuation(true) }
func (r *task[T]) AsAsync() Task[T]                        { return r._async() }
func (r *task[T]) IsAsync() bool                           { return isAsync(r.status) }

func (r *task[T]) lock() (previous uint32) {
	for {
		current := atomic.LoadUint32(&r.status)

		if isDone(current) {
			return current
		}

		if !isLocked(current) && atomic.CompareAndSwapUint32(&r.status, current, current|_LOCKED) {
			return current
		}
	}
}

func (r *task[T]) unlock(status uint32) {
	atomic.StoreUint32(&r.status, status)
}

func (r *task[T]) completeOrFail(val T, err error) bool {
	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return false
	}

	fn, ch := r.getAnyHandlerUnsafe(current)

	updated := r.encodeValOrErrAndStoreUnsafe(val, err)

	r.unlock(updated)

	if ch != nil {
		close(ch)
	}

	if fn != nil {
		r.executeHandlerWithValOrErr(val, err, fn, isAsync(current))
	}

	return true
}

func (r *task[T]) getOrAddDoneChan() <-chan x.Void {
	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return x.ClosedChan
	}

	ch := make(chan x.Void)
	r._addHandlerNoReturn(func(_ T, _ error) {
		close(ch)
	})

	r.handlerStackPushUnsafe(current, func(_ T, _ error) {
		close(ch)
	}, true)

	r.unlock(current | _FN)

	return ch
}

func (r *task[T]) addFnHandler(fn Handler[T], noTask bool) *task[T] {
	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return r.executeHandlerUnsafe(current, fn)
	}

	task := r.handlerStackPushUnsafe(current, fn, noTask)

	r.unlock(current | _FN)

	return task
}

func (r *task[T]) loadValueOrErrorUnsafe(status uint32) (value T, err error) {
	if hasError(status) {
		//err = *(*error)(atomic.LoadPointer(&r.taskState.state))
		err = *(*error)(r.state)
		return
	}

	//ptr := (*T)(atomic.LoadPointer(&r.taskState.state))
	ptr := (*T)(r.state)
	if ptr != nil {
		return *ptr, nil
	}

	return
}

func (r *task[T]) setEitherNoReturn(value T, err error) {
	r.completeOrFail(value, err)
}

func (r *task[T]) handlerStackPushUnsafe(current uint32, fn Handler[T], noTask bool) *task[T] {
	if !hasHandler(current) {
		// first fn, so just encode and taskState
		r.encodeFnAndStoreUnsafe(fn)
		return r
	}

	task, h := r.getCombinedHandler(fn, r.getHandlerUnsafe(), noTask)
	r.encodeFnAndStoreUnsafe(h)

	return task
}

func (r *task[T]) getCombinedHandler(newer Handler[T], older Handler[T], noTask bool) (*task[T], Handler[T]) {
	// already have a handler, so wrap the current one and the new one
	if noTask {
		h := &onSuccessOrErrorHandler[T]{older: older, newer: newer}
		return r, h.invoke
	}

	task := &task[T]{_FN, unsafe.Pointer(&newer)}
	h := onSuccessOrErrorTaskHandler[T]{task, older}

	return task, h.invoke
}

func (r *task[T]) executeHandlerUnsafe(current uint32, fn Handler[T]) *task[T] {
	v, err := r.loadValueOrErrorUnsafe(current)

	r.executeHandlerWithValOrErr(v, err, fn, isAsync(current))

	return r
}

func (r *task[T]) executeHandlerWithValOrErr(val T, err error, fn Handler[T], async bool) {
	if !async {
		fn(val, err)
	} else {
		go fn(val, err)
	}
}

func (r *task[T]) getAnyHandlerUnsafe(current uint32) (Handler[T], chan x.Void) {
	if hasHandler(current) {
		return r.getHandlerUnsafe(), nil
	}

	return nil, nil
}

func (r *task[T]) getHandlerUnsafe() Handler[T] {
	return *(*Handler[T])(r.state)
	//return *(*Handler[T])(atomic.LoadPointer(&r.taskState.state))
}

func (r *task[T]) encodeToBlockChanAndStoreUnsafe(bkh *onDoneChanBlockRunHandler[T]) {
	atomic.StorePointer(&r.state, unsafe.Pointer(&bkh))
}

func (r *task[T]) encodeFnAndStoreUnsafe(fn Handler[T]) {
	atomic.StorePointer(&r.state, unsafe.Pointer(&fn))
}

func (r *task[T]) encodeValOrErrAndStoreUnsafe(val T, err error) uint32 {
	if err == nil {
		atomic.StorePointer(&r.state, unsafe.Pointer(&val))
		return _VALUE
	}

	atomic.StorePointer(&r.state, unsafe.Pointer(&err))

	return utils.IfElse(err == x.ErrOperationCancelled, _CANCELLED_OR_ERROR, _ERROR)
}
