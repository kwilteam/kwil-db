package rx

import (
	"kwil/x/utils"
	"sync/atomic"
	"unsafe"
)

type taskState struct {
	status uint32
	state  unsafe.Pointer
}

func (r *Task[T]) lock() (previous uint32) {
	for {
		current := atomic.LoadUint32(&r.store.status)

		if isDone(current) {
			return current
		}

		if !isLocked(current) && atomic.CompareAndSwapUint32(&r.store.status, current, current|_LOCKED) {
			return current
		}
	}
}

func (r *Task[T]) unlock(status uint32) {
	atomic.StoreUint32(&r.store.status, status)
}

func (r *Task[T]) completeOrFail(val T, err error) bool {
	if isCompletedOrigin(r.store.status) {
		return false
	}

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

func (r *Task[T]) addFnHandler(fn Handler[T], noTask bool) *Task[T] {
	if isCompletedOrigin(r.store.status) {
		return r.executeHandlerUnsafe(r.store.status, fn)
	}

	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return r.executeHandlerUnsafe(current, fn)
	}

	task := r.handlerStackPushUnsafe(current, fn, noTask)

	r.unlock(current | _FN)

	return task
}

func (r *Task[T]) getOrAddDoneBlockChanHandler() chan Void {
	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return _closedChan
	}

	var bkh *onDoneChanBlockRunHandler[T]
	if hasBlockingDoneHandler(current) {
		bkh = r.getDoneChanBlockHandlerUnsafe()
	} else {
		bkh = r.getAndEncodeNewDoneChanBlockRunHandlerUnsafe(current)
	}

	r.unlock(current | _DONE_BLOCKING_HANDLER)

	return bkh.chDone
}

func (r *Task[T]) loadValueOrErrorUnsafe(status uint32) (value T, err error) {
	if hasError(status) {
		//err = *(*error)(atomic.LoadPointer(&r.taskState.state))
		err = *(*error)(r.store.state)
		return
	}

	//ptr := (*T)(atomic.LoadPointer(&r.taskState.state))
	ptr := (*T)(r.store.state)
	if ptr != nil {
		return *ptr, nil
	}

	return
}

func (r *Task[T]) setEitherNoReturn(value T, err error) {
	r.completeOrFail(value, err)
}

func (r *Task[T]) handlerStackPushUnsafe(current uint32, fn Handler[T], noTask bool) *Task[T] {
	if !hasAnyHandler(current) {
		// first fn, so just encode and taskState
		r.encodeFnAndStoreUnsafe(fn)
		return r
	}

	if !hasBlockingDoneHandler(current) {
		task, h := r.getCombinedHandler(fn, r.getHandlerUnsafe(), noTask)
		r.encodeFnAndStoreUnsafe(h)
		return task
	}

	bkh := r.getDoneChanBlockHandlerUnsafe()
	if bkh.fn == nil {
		bkh.fn = fn
		return r
	}

	task, h := r.getCombinedHandler(fn, bkh.fn, noTask)
	bkh.fn = h

	return task
}

func (r *Task[T]) getCombinedHandler(newer Handler[T], older Handler[T], noTask bool) (*Task[T], Handler[T]) {
	// already have a handler, so wrap the current one and the new one
	if noTask {
		h := &onSuccessOrErrorHandler[T]{older: older, newer: newer}
		return r, h.invoke
	}

	task := &Task[T]{&taskState{
		status: _FN,
		state:  unsafe.Pointer(&newer),
	}}
	h := onSuccessOrErrorTaskHandler[T]{task, older}

	return task, h.invoke
}

func (r *Task[T]) getAndEncodeNewDoneChanBlockRunHandlerUnsafe(current uint32) *onDoneChanBlockRunHandler[T] {
	var fn Handler[T]
	if hasHandler(current) {
		fn = r.getHandlerUnsafe()
	}

	bkh := &onDoneChanBlockRunHandler[T]{chDone: make(chan Void), fn: fn}
	r.encodeToBlockChanAndStoreUnsafe(bkh)

	return bkh
}

func (r *Task[T]) executeHandlerUnsafe(current uint32, fn Handler[T]) *Task[T] {
	v, err := r.loadValueOrErrorUnsafe(current)

	r.executeHandlerWithValOrErr(v, err, fn, isAsync(current))

	return r
}

func (r *Task[T]) executeHandlerWithValOrErr(val T, err error, fn Handler[T], async bool) {
	if !async {
		fn(val, err)
	} else {
		go fn(val, err)
	}
}

func (r *Task[T]) getAnyHandlerUnsafe(current uint32) (Handler[T], chan Void) {
	if hasBlockingDoneHandler(current) {
		bkh := r.getDoneChanBlockHandlerUnsafe()
		return bkh.invoke, bkh.chDone
	}

	if hasHandler(current) {
		return r.getHandlerUnsafe(), nil
	}

	return nil, nil
}

func (r *Task[T]) getHandlerUnsafe() Handler[T] {
	return *(*Handler[T])(r.store.state)
	//return *(*Handler[T])(atomic.LoadPointer(&r.taskState.state))
}

func (r *Task[T]) getDoneChanBlockHandlerUnsafe() *onDoneChanBlockRunHandler[T] {
	return (*onDoneChanBlockRunHandler[T])(r.store.state)
	//return (*onDoneChanBlockRunHandler[T])(atomic.LoadPointer(&r.taskState.state))
}

func (r *Task[T]) encodeToBlockChanAndStoreUnsafe(bkh *onDoneChanBlockRunHandler[T]) {
	//*r.state = unsafe.Pointer(bkh)
	atomic.StorePointer(&r.store.state, unsafe.Pointer(&bkh))
}

func (r *Task[T]) encodeFnAndStoreUnsafe(fn Handler[T]) {
	//*r.state = unsafe.Pointer(&fn)
	atomic.StorePointer(&r.store.state, unsafe.Pointer(&fn))
}

func (r *Task[T]) encodeValOrErrAndStoreUnsafe(val T, err error) uint32 {
	if err == nil {
		//*r.state = unsafe.Pointer(&val)
		atomic.StorePointer(&r.store.state, unsafe.Pointer(&val))
		return _VALUE
	}

	//*r.state = unsafe.Pointer(&err)
	atomic.StorePointer(&r.store.state, unsafe.Pointer(&err))

	return utils.IfElse(err == ErrCancelled, _CANCELLED_OR_ERROR, _ERROR)
}
