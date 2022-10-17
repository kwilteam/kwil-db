package async

import (
	"context"
	. "kwil/x"
	"kwil/x/errx"
	"kwil/x/utils"
	"sync/atomic"
	"unsafe"
)

type _task[T any] struct {
	status uint32
	state  unsafe.Pointer
}

func (r *_task[T]) GetError() error                              { return r._getError() }
func (r *_task[T]) Get() T                                       { return r._get() }
func (r *_task[T]) IsError() bool                                { return r._isError() }
func (r *_task[T]) IsCancelled() bool                            { return r._isCancelled() }
func (r *_task[T]) IsErrorOrCancelled() bool                     { return r._isErrorOrCancelled() }
func (r *_task[T]) IsDone() bool                                 { return r._isDone() }
func (r *_task[T]) DoneCh() <-chan Void                          { return r._doneChan() }
func (r *_task[T]) Fail(err error) bool                          { return r._fail(err) }
func (r *_task[T]) Complete(value T) bool                        { return r._complete(value) }
func (r *_task[T]) CompleteOrFail(value T, err error) bool       { return r._completeOrFail(value, err) }
func (r *_task[T]) Cancel() bool                                 { return r._cancel() }
func (r *_task[T]) GetOrError() (T, error)                       { return r._getOrError() }
func (r *_task[T]) Await(ctx context.Context) (ok bool)          { return r._await(ctx) }
func (r *_task[T]) Then(fn func(T)) Task[T]                      { return r._then(fn) }
func (r *_task[T]) ThenCh(ch chan T) Task[T]                     { return r._thenCh(ch) }
func (r *_task[T]) Catch(fn func(error)) Task[T]                 { return r._catch(fn) }
func (r *_task[T]) CatchCh(ch chan error) Task[T]                { return r._catchCh(ch) }
func (r *_task[T]) Handle(fn func(T, error) (T, error)) Task[T]  { return r._handle(fn) }
func (r *_task[T]) Compose(fn func(T, error) Task[T]) Task[T]    { return r._compose(fn) }
func (r *_task[T]) ThenCatchFinally(fn *Continuation[T]) Task[T] { return r._whenComplete(fn.invoke) }
func (r *_task[T]) WhenComplete(fn func(T, error)) Task[T]       { return r._whenComplete(fn) }
func (r *_task[T]) WhenCompleteCh(ch chan *Result[T]) Task[T]    { return r._whenCompleteCh(ch) }
func (r *_task[T]) OnComplete(fn *Continuation[T])               { r._whenComplete(fn.invoke) }
func (r *_task[T]) AsAction() Action                             { return r._asAction() }
func (r *_task[T]) AsListenable() Listenable[T]                  { return r }
func (r *_task[T]) AsAsync(e Executor) Task[T]                   { return r._async(e) }
func (r *_task[T]) IsAsync() bool                                { return isAsync(r.status) }

func (r *_task[T]) lock() (previous uint32) {
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

func (r *_task[T]) unlock(status uint32) {
	atomic.StoreUint32(&r.status, status)
}

func (r *_task[T]) completeOrFailNoReturn(val T, err error) {
	r.completeOrFail(val, err)
}

func (r *_task[T]) completeOrFail(val T, err error) bool {
	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return false
	}

	fn, ch := r.getAnyHandler(current)

	updated := r.encodeValOrErrAndStore(val, err)

	r.unlock(updated)

	if ch != nil {
		close(ch)
	}

	if fn != nil {
		r.executeHandlerWithValOrErr(val, err, fn, isAsync(current))
	}

	return true
}

func (r *_task[T]) getOrAddDoneChan() <-chan Void {
	ch := make(chan Void)
	r.addFnHandler(func(_ T, _ error) {
		close(ch)
	}, true)

	return ch
}

func (r *_task[T]) addFnHandler(fn func(T, error), noTask bool) *_task[T] {
	current := r.lock()
	if isDone(current) {
		r.unlock(current)
		return r.executeHandler(current, fn)
	}

	task := r.handlerStackPush(current, fn, noTask)

	r.unlock(current | _FN)

	return task
}

func (r *_task[T]) loadValueOrError(status uint32) (value T, err error) {
	if hasError(status) {
		err = *(*error)(atomic.LoadPointer(&r.state))
		return
	}

	ptr := (*T)(atomic.LoadPointer(&r.state))
	if ptr != nil {
		return *ptr, nil
	}

	return
}

func (r *_task[T]) setEitherNoReturn(value T, err error) {
	r.completeOrFail(value, err)
}

func (r *_task[T]) handlerStackPush(current uint32, fn func(T, error), noTask bool) *_task[T] {
	if !hasHandler(current) {
		// first fn, so just set it
		r.encodeFnAndStore(fn)
		return r
	}

	task, h := r.getCombinedHandler(fn, r.getHandler(), noTask)
	r.encodeFnAndStore(h)

	return task
}

func (r *_task[T]) getCombinedHandler(newer func(T, error), older func(T, error), noTask bool) (*_task[T], func(T, error)) {
	// already have a handler, so wrap the current one and the new one
	if noTask {
		h := &onCompleteStackNodeHandler[T]{next: older, fn: newer}
		return r, h.invoke
	}

	task := &_task[T]{_FN, unsafe.Pointer(&newer)}
	h := onSuccessOrErrorTaskHandler[T]{task, older}

	return task, h.invoke
}

func (r *_task[T]) executeHandler(current uint32, fn func(T, error)) *_task[T] {
	v, err := r.loadValueOrError(current)

	r.executeHandlerWithValOrErr(v, err, fn, isAsync(current))

	return r
}

func (r *_task[T]) executeHandlerWithValOrErr(val T, err error, fn func(T, error), async bool) {
	if !async {
		fn(val, err)
	} else {
		go fn(val, err)
	}
}

func (r *_task[T]) getAnyHandler(current uint32) (func(T, error), chan Void) {
	if hasHandler(current) {
		return r.getHandler(), nil
	}

	return nil, nil
}

func (r *_task[T]) getHandler() func(T, error) {
	return *(*func(T, error))(atomic.LoadPointer(&r.state))
}

func (r *_task[T]) encodeFnAndStore(fn func(T, error)) {
	atomic.StorePointer(&r.state, unsafe.Pointer(&fn))
}

func (r *_task[T]) encodeValOrErrAndStore(val T, err error) uint32 {
	if err == nil {
		atomic.StorePointer(&r.state, unsafe.Pointer(&val))
		return _VALUE
	}

	atomic.StorePointer(&r.state, unsafe.Pointer(&err))

	return utils.IfElse(errx.IsCancelled(err), _CANCELLED_OR_ERROR, _ERROR)
}
