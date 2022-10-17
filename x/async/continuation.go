package async

type Continuation[T any] struct {
	Then    func(T)
	Catch   func(error)
	Finally func()
}

func (c *Continuation[T]) invoke(value T, err error) {
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
