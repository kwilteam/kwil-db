package backend

import (
	"ksl"
	"sync"
)

var connectors sync.Map
var defaultConnector Connector
var defmu sync.Mutex

type Connector interface {
	Name() string
	MaxIdentifierLength() uint
	ScalarTypeForNativeType(t ksl.Type) ksl.BuiltInScalar
	DefaultNativeTypeForScalar(t ksl.BuiltInScalar) ksl.Type
	ParseNativeType(name string, args ...string) (ksl.Type, error)
}

func Get(name string) Connector {
	if c, ok := connectors.Load(name); ok {
		return c.(Connector)
	}

	if defaultConnector != nil {
		return defaultConnector
	}

	panic("connector: no connector registered for " + name)
}

func Register(c Connector) {
	if c == nil {
		panic("connector: Register connector is nil")
	}

	if _, ok := connectors.Load(c.Name()); ok {
		panic("connector: Register called twice for connector " + c.Name())
	}

	connectors.Store(c.Name(), c)
}

func RegisterDefault(c Connector) {
	if c == nil {
		panic("connector: RegisterDefault connector is nil")
	}

	defmu.Lock()
	defer defmu.Unlock()

	if defaultConnector != nil {
		panic("connector: RegisterDefault called twice")
	}

	defaultConnector = c
}
