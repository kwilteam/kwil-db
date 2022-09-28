package rx

import (
	"fmt"
	"unsafe"
)

// NewTask creates a new task that will execution
// continuations synchronously
func NewTask[T any]() *Task[T] {
	status := uint32(0)
	return &Task[T]{status: &status}
}

// NewTaskAsync creates a new task that will initiate
// execution of continuations asynchronously
func NewTaskAsync[T any]() *Task[T] {
	status := _ASYNC_CONTINUATIONS
	return &Task[T]{status: &status}
}

// Success returns a completed Task with the param 'value'.
func Success[T any](value T) *Task[T] {
	return &Task[T]{state: unsafe.Pointer(&value), status: _taskValuePtr}
}

// Failure returns an errored Task with the param 'err'.
func Failure[T any](err error) *Task[T] {
	return &Task[T]{state: unsafe.Pointer(&err), status: _taskErrorPtr}
}

// FailureC will return a new Continuation that is in a completed failed state
func FailureC(err error) *Continuation {
	return Failure[struct{}](err).AsContinuation()
}

// SuccessC will return a new Continuation that is in a completed successful state
func SuccessC() *Continuation {
	return Success(struct{}{}).AsContinuation()
}

func Exec(fn func() error) *Continuation {
	task := NewTask[struct{}]()

	go func(f func() error, t *Task[struct{}]) {
		if t.IsDone() {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if !ok {
					e = fmt.Errorf("unknown panic: %v", r)
				}
				t.Fail(e)
			}
		}()

		t.CompleteOrFail(struct{}{}, f())
	}(fn, task)

	return &Continuation{task}
}

func Invoke[T any](fn func() T) *Task[T] {
	task := NewTask[T]()

	go func(f func() T, t *Task[T]) {
		if t.IsDone() {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if !ok {
					e = fmt.Errorf("unknown panic: %v", r)
				}
				t.Fail(e)
			}
		}()

		t.Complete(f())
	}(fn, task)

	return task
}

func Call[T any](fn func() (T, error)) *Task[T] {
	task := NewTask[T]()

	go func(f func() (T, error), t *Task[T]) {
		if t.IsDone() {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if !ok {
					e = fmt.Errorf("unknown panic: %v", r)
				}
				t.Fail(e)
			}
		}()

		val, err := f()
		t.CompleteOrFail(val, err)
	}(fn, task)

	return task
}

func InvokeWithArgs[T, U any](args U, fn func(U) T) *Task[T] {
	task := NewTask[T]()

	go func(a U, f func(U) T, t *Task[T]) {
		if t.IsDone() {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if !ok {
					e = fmt.Errorf("unknown panic: %v", r)
				}
				t.Fail(e)
			}
		}()

		t.Complete(f(a))
	}(args, fn, task)

	return task
}

func CallWithArgs[T, U any](args U, fn func(U) (T, error)) *Task[T] {
	task := NewTask[T]()

	go func(a U, f func(U) (T, error), t *Task[T]) {
		if t.IsDone() {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if !ok {
					e = fmt.Errorf("unknown panic: %v", r)
				}
				t.Fail(e)
			}
		}()

		val, err := f(a)
		t.CompleteOrFail(val, err)
	}(args, fn, task)

	return task
}
