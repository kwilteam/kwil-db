package async

import (
	. "kwil/x"
)

// The below are internally used to adapt handlers to a common
// callback type Methods of the structs are used as function
// pointers with the handler callbacks from a _task
// TODO: in the future, move to using interfaces for simpler readability
// over any potential/unlikely meaningful perf gains

type continuationRequestAdapter struct {
	fn func(error)
}

func (r *continuationRequestAdapter) invoke(_ Void, err error) {
	r.fn(err)
}

type onSuccessContinuationHandler struct {
	fn func()
}

func (h *onSuccessContinuationHandler) invoke(_ Void, err error) {
	if err == nil {
		h.fn()
	}
}

type onSuccessOrErrorTaskHandler[T any] struct {
	task *_task[T]
	fn   func(T, error)
}

func (h *onSuccessOrErrorTaskHandler[T]) invoke(value T, err error) {
	h.task.CompleteOrFail(value, err)
	h.fn(value, err)
}

type onCompleteStackNodeHandler[T any] struct {
	next func(T, error)
	fn   func(T, error)
}

func (h *onCompleteStackNodeHandler[T]) invoke(value T, err error) {
	h.fn(value, err)
	h.next(value, err)
}

type onSuccessHandler[T any] struct {
	fn func(T)
}

func (h *onSuccessHandler[T]) invoke(value T, err error) {
	if err == nil {
		h.fn(value)
	}
}

type onErrorHandler[T any] struct {
	fn func(error)
}

func (h *onErrorHandler[T]) invoke(_ T, err error) {
	if err != nil {
		h.fn(err)
	}
}

type onDoneChanBlockRunHandler[T any] struct {
	chDone chan Void
	fn     func(T, error)
}

func (h *onDoneChanBlockRunHandler[T]) invoke(v T, err error) {
	close(h.chDone)
	h.fn(v, err)
}

type onCompose[T any] struct {
	task *_task[T]
	fn   func(T, error) Task[T]
}

func (h *onCompose[T]) invoke(v T, err error) {
	h.fn(v, err).WhenComplete(h.fn_complete)
}

func (h *onCompose[T]) fn_complete(v T, e error) {
	h.task.completeOrFailNoReturn(v, e)
}

type onHandle[T any] struct {
	task *_task[T]
	fn   func(T, error) (T, error)
}

func (h *onHandle[T]) invoke(v T, e error) {
	h.task.completeOrFailNoReturn(h.fn(v, e))
}

type onAsyncHandler[T any] struct {
	task *_task[T]
}

func (h *onAsyncHandler[T]) invoke(v T, e error) {
	if !h.task.IsDone() {
		go h.task.completeOrFail(v, e)
	}
}

type asActionHandler[T any] struct {
	action *action
}

func (h *asActionHandler[T]) invoke(_ T, err error) {
	h.action.CompleteOrFail(err)
}

type executorHandler[T any] struct {
	task *_task[T]
	e    Executor
	err  error
	v    T
}

func (h *executorHandler[T]) run() {
	h.task.CompleteOrFail(h.v, h.err)
}

func (h *executorHandler[T]) invoke(v T, err error) {
	h.err = err
	h.v = v
	h.e.Execute(h.run)
}
