package rx

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

func (r *continuationRequestAdapter) invoke(_ Void, err error) {
	r.fn(err)
}

type onSuccessContinuationHandler struct {
	fn Runnable
}

func (h *onSuccessContinuationHandler) invoke(_ Void, err error) {
	if err == nil {
		h.fn()
	}
}

type continuationRequestAdapterAsync struct {
	fn func(error)
}

func (r *continuationRequestAdapterAsync) invoke(_ Void, err error) {
	go r.fn(err)
}

type onSuccessContinuationHandlerAsync struct {
	fn Runnable
}

func (h *onSuccessContinuationHandlerAsync) invoke(_ Void, err error) {
	if err == nil {
		go h.fn()
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

type onSuccessOrErrorHandler[T any] struct {
	older Handler[T]
	newer Handler[T]
}

func (h *onSuccessOrErrorHandler[T]) invoke(value T, err error) {
	h.newer(value, err)
	h.older(value, err)
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

type onCompleteRunHandler[T any] struct {
	fn Runnable
}

func (h *onCompleteRunHandler[T]) invoke(_ T, _ error) {
	h.fn()
}

type asContinuationHandler[T any] struct {
	c *Continuation
}

func (h *asContinuationHandler[T]) invoke(_ T, err error) {
	h.c.task.CompleteOrFail(Void{}, err)
}

type onDoneChanBlockRunHandler[T any] struct {
	chDone chan Void
	fn     Handler[T]
}

func (h *onDoneChanBlockRunHandler[T]) invoke(v T, err error) {
	close(h.chDone)
	h.fn(v, err)
}

type onSuccessHandlerAsync[T any] struct {
	fn ValueHandler[T]
}

func (h *onSuccessHandlerAsync[T]) invoke(value T, err error) {
	if err == nil {
		go h.fn(value)
	}
}

type onErrorHandlerAsync[T any] struct {
	fn ErrorHandler
}

func (h *onErrorHandlerAsync[T]) invoke(_ T, err error) {
	if err != nil {
		go h.fn(err)
	}
}

type onCompleteRunHandlerAsync[T any] struct {
	fn Runnable
}

func (h *onCompleteRunHandlerAsync[T]) invoke(_ T, _ error) {
	go h.fn()
}

type onCompleteHandlerAsync[T any] struct {
	fn Handler[T]
}

func (h *onCompleteHandlerAsync[T]) invoke(val T, err error) {
	go h.fn(val, err)
}
