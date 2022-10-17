package async

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

// CompletedTask returns a completed Task with the param 'value'.
func CompletedTask[T any](value T) Task[T] {
	return &task_value[T]{value}
}

// FailedTask returns an errored Task with the param 'err'.
func FailedTask[T any](err error) Task[T] {
	return &task_error[T]{err}
}

// CancelledTask returns a cancelled Task. Equivalent
// to an errored task containing a context.CancelledTask
func CancelledTask[T any]() Task[T] {
	return &task_error[T]{errx.ErrOperationCancelled()}
}

// FailedAction will return a new Action that is in a
// completed failed state
func FailedAction(err error) Action {
	return &action_err{err}
}

// CompletedAction will return a new Action that is in a
// completed successful state
func CompletedAction() Action {
	return &action_value{}
}

// CancelledAction returns a cancelled Task. Equivalent
// to an errored task containing a context.CancelledTask
func CancelledAction() Action {
	return &action_err{errx.ErrOperationCancelled()}
}
