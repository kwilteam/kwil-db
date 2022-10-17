package x

var immediateExecutor Executor = &immediate_executor{}
var asyncExecutor Executor = &async_executor{}

// AsyncExecutor returns an executor that will execute the given
// function asynchronously using the native go scheduler (e.g., go func(){}).
func AsyncExecutor() Executor {
	return asyncExecutor
}

// ImmediateExecutor will execute the given function immediately on the
// callers thread.
func ImmediateExecutor() Executor {
	return immediateExecutor
}

type async_executor struct{}

func (e *async_executor) Execute(fn Runnable) {
	go fn()
}

type immediate_executor struct{}

func (e *immediate_executor) Execute(fn Runnable) {
	fn()
}
