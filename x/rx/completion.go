package rx

type Completion[T any] struct {
	Then    func(T)
	Catch   func(error)
	Finally func()
}

func (c *Completion[T]) Invoke(value T, err error) {
	c._invoke(value, err)
}

func (c *Completion[T]) _invoke(value T, err error) {
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

type CompletionC struct {
	Then    func()
	Catch   func(error)
	Finally func()
}

func (c *CompletionC) Invoke(err error) {
	c._invoke(err)
}

func (c *CompletionC) _invoke(err error) {
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
