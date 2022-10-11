package rx

import (
	"kwil/x/errx"
)

// NewTask creates a new Task that will execute
// continuations synchronously
func NewTask[T any]() Task[T] {
	return newTask[T]()
}

// NewTaskAsync creates a new Task that will execute
// continuations synchronously
func NewTaskAsync[T any]() Task[T] {
	return newTaskAsync[T]()
}

// NewAction creates a new Task that will execute
// continuations synchronously
func NewAction() Action {
	return _newAction()
}

// NewActionAsync creates a new Task that will execute
// continuations synchronously
func NewActionAsync() Action {
	return _newActionAsync()
}

// Success returns a completed Task with the param 'value'.
func Success[T any](value T) Task[T] {
	return &task_value[T]{value}
}

// Failure returns an errored Task with the param 'err'.
func Failure[T any](err error) Task[T] {
	return &task_error[T]{err}
}

// Cancelled returns a cancelled Task. Equivalent
// to an errored task containing a context.Cancelled
func Cancelled[T any]() Task[T] {
	//l := context.Canceled
	return &task_error[T]{errx.ErrOperationCancelled()}
}

// FailureA will return a new Action that is in a
// completed failed state
func FailureA(err error) Action {
	return &action_err{err}
}

// SuccessA will return a new Action that is in a
// completed successful state
func SuccessA() Action {
	return &action_value{}
}

// CancelledA returns a cancelled Task. Equivalent
// to an errored task containing a context.Cancelled
func CancelledA() Action {
	return &action_err{errx.ErrOperationCancelled()}
}
