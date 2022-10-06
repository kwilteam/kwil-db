package rx

type Completion[T any] struct {
	Then    func(T)
	Catch   func(error)
	Finally func()
	Async   bool
}

func (c *Completion[T]) Invoke(value T, err error) {
	if c.Async {
		go c._invoke(value, err)
	} else {
		c._invoke(value, err)
	}
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
	Async   bool
}

func (c *CompletionC) Invoke(err error) {
	if c.Async {
		go c._invoke(err)
	} else {
		c._invoke(err)
	}
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
