package driver

type InjectableVar struct {
	Name       string
	DefaultVal any
}

func (c *Connection) listInjectables() []string {
	var injectables []string
	for _, injectable := range c.injectables {
		injectables = append(injectables, injectable.Name)
	}
	return injectables
}

func (c *Connection) addInjectables(args map[string]interface{}) {
	for _, injectable := range c.injectables {
		if _, ok := args[injectable.Name]; !ok {
			args[injectable.Name] = injectable.DefaultVal
		}
	}
}
