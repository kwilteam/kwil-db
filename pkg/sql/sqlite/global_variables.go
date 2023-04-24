package sqlite

type GlobalVariable struct {
	Name     string
	Default  any
	DataType DataType
}

func (c *Connection) containsGlobalVar(name string) bool {
	for _, v := range c.globalVariables {
		if v.Name == name {
			return true
		}
	}

	return false
}
