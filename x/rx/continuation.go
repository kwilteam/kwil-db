package rx

type Executor interface {
	Execute(fn func())
}

type async_executor struct{}

func (e *async_executor) Execute(fn func()) {
	go fn()
}

var asyncExecutor Executor = &async_executor{}

func AsyncExecutor() Executor {
	return asyncExecutor
}

type immediate_executor struct{}

func (e *immediate_executor) Execute(fn func()) {
	fn()
}

var immediateExecutor Executor = &immediate_executor{}

func ImmediateExecutor() Executor {
	return immediateExecutor
}

type Continuation struct {
	next *Continuation
	run  func()
	done bool
}

type ContinuationT[T any] struct {
	Then    func(T)
	Catch   func(error)
	Finally func()
}

func (c *ContinuationT[T]) invoke(value T, err error) {
	if err == nil {
		if c.Then != nil {
			c.Then(value)
		}
	} else {
		if c.Catch != nil {
			c.Catch(err)
		}
	}

	if c.Finally != nil {
		c.Finally()
	}
}

type ContinuationA struct {
	Then    func()
	Catch   func(error)
	Finally func()
}

func (c *ContinuationA) invoke(err error) {
	if err == nil {
		if c.Then != nil {
			c.Then()
		}
	} else {
		if c.Catch != nil {
			c.Catch(err)
		}
	}

	if c.Finally != nil {
		c.Finally()
	}
}
