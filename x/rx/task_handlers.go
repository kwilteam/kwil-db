package rx

import "kwil/x"

type ArgsHandler[T, U any] func(args U, value T, err error)
type Handler[T any] func(T, error)
type ValueHandler[T any] func(T)
type ErrorHandler func(error)
type Runnable func()

// The below are internally used to adapt handlers to a common
// callback type Methods of the structs are used as function
// pointers with the handler callbacks from a task
// TODO: in the future, move to using interfaces for simpler readability
// over any potential/unlikely meaningful perf gains

type continuationRequestAdapter struct {
	fn func(error)
}

func (r *continuationRequestAdapter) invoke(_ x.Void, err error) {
	r.fn(err)
}

type onSuccessContinuationHandler struct {
	fn Runnable
}

func (h *onSuccessContinuationHandler) invoke(_ x.Void, err error) {
	if err == nil {
		h.fn()
	}
}

type onSuccessOrErrorTaskHandler[T any] struct {
	task *task[T]
	fn   Handler[T]
}

func (h *onSuccessOrErrorTaskHandler[T]) invoke(value T, err error) {
	h.task.CompleteOrFail(value, err)
	h.fn(value, err)
}

type onCompleteStackNodeHandler[T any] struct {
	next Handler[T]
	fn   Handler[T]
}

func (h *onCompleteStackNodeHandler[T]) invoke(value T, err error) {
	h.fn(value, err)
	h.next(value, err)
}

type onSuccessHandler[T any] struct {
	fn ValueHandler[T]
}

func (h *onSuccessHandler[T]) invoke(value T, err error) {
	if err == nil {
		h.fn(value)
	}
}

type onErrorHandler[T any] struct {
	fn ErrorHandler
}

func (h *onErrorHandler[T]) invoke(_ T, err error) {
	if err != nil {
		h.fn(err)
	}
}

type onDoneChanBlockRunHandler[T any] struct {
	chDone chan x.Void
	fn     Handler[T]
}

func (h *onDoneChanBlockRunHandler[T]) invoke(v T, err error) {
	close(h.chDone)
	h.fn(v, err)
}

type onCompose[T any] struct {
	task Task[T]
	fn   func(T, error) Task[T]
}

func (h *onCompose[T]) invoke(v T, err error) {
	h.fn(v, err).WhenComplete(h.fn_complete)
}

func (h *onCompose[T]) fn_complete(v T, e error) {
	h.task.CompleteOrFail(v, e)
}

type onHandle[T any] struct {
	task Task[T]
	fn   func(T, error) (T, error)
}

func (h *onHandle[T]) invoke(v T, e error) {
	h.task.CompleteOrFail(h.fn(v, e))
}
