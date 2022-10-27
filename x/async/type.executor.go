package async

import "kwil/x"

type Job x.Runnable

type Executor interface {
	Execute(Job)
}

var immediateExecutor Executor = &immediate_executor{}
var asyncExecutor Executor = &async_executor{}

// DefaultExecutor returns an executor that will execute the given
// function asynchronously using the native go scheduler (e.g., go func(){}).
func DefaultExecutor() Executor {
	return asyncExecutor
}

// ImmediateExecutor will execute the given function immediately on the
// callers thread.
func ImmediateExecutor() Executor {
	return immediateExecutor
}

type async_executor struct{}

func (e *async_executor) Execute(job Job) {
	go job()
}

type immediate_executor struct{}

func (e *immediate_executor) Execute(job Job) {
	job()
}
