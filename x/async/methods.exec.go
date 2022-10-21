package async

import (
	"fmt"
)

// Call executes the given function asynchronously in a go routine. A Task
// is returned either completed with the result of the function or an error
// if the function returns an error.
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

// CallWithArgs has all the same behavior as Call, but additionally pass an argument
// to the function.
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
