package driver

type ConnOpt func(*Connection)

func WithPath(path string) ConnOpt {
	return func(c *Connection) {
		c.path = path
	}
}

func ReadOnly() ConnOpt {
	return func(c *Connection) {
		c.readOnly = true
	}
}
