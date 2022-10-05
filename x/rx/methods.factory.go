package rx

import (
	"fmt"
	"unsafe"
)

// NewTask creates a new task that will execute
// continuations synchronously
func NewTask[T any]() Task[T] {
	return newTask[T]()
}

// NewTaskAsync creates a new task that will execute
// continuations synchronously
func NewTaskAsync[T any]() Task[T] {
	return newTaskAsync[T]()
}

// NewContinuation creates a new task that will execute
// continuations synchronously
func NewContinuation() *Continuation {
	return &Continuation{newTask[Void]()}
}

// NewContinuationAsync creates a new task that will execute
// continuations synchronously
func NewContinuationAsync() *Continuation {
	return &Continuation{newTaskAsync[Void]()}
}

// Success returns a completed task with the param 'value'.
func Success[T any](value T) Task[T] {
	return &task[T]{
		status: _VALUE,
		state:  unsafe.Pointer(&value),
	}
}

// Failure returns an errored task with the param 'err'.
func Failure[T any](err error) Task[T] {
	status := _ERROR
	if err == ErrCancelled {
		status = _CANCELLED_OR_ERROR
	}
	return &task[T]{
		status: status,
		state:  unsafe.Pointer(&err),
	}
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
	task := newTask[Void]()

	go func(f func() error, t Task[Void]) {
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

		t.CompleteOrFail(Void{}, f())
	}(fn, task)

	return &Continuation{task}
}

func Invoke[T any](fn func() T) Task[T] {
	task := NewTask[T]()

	go func(f func() T, t Task[T]) {
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

func Call[T any](fn func() (T, error)) Task[T] {
	task := NewTask[T]()

	go func(f func() (T, error), t Task[T]) {
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

func InvokeWithArgs[T, U any](args U, fn func(U) T) Task[T] {
	task := NewTask[T]()

	go func(a U, f func(U) T, t Task[T]) {
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

func CallWithArgs[T, U any](args U, fn func(U) (T, error)) Task[T] {
	task := NewTask[T]()

	go func(a U, f func(U) (T, error), t Task[T]) {
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
