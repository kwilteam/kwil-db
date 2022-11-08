package kslspec

import (
	"fmt"
	"log"
)

type TypeRegistry struct {
	types   []*TypeSpec
	aliases map[string]*TypeSpec

	spec   func(Type) (*ConcreteType, error)
	parser func(string) (Type, error)
	format func(*ConcreteType) (string, error)
}

func NewRegistry(opts ...TypeRegistryOption) *TypeRegistry {
	r := &TypeRegistry{aliases: make(map[string]*TypeSpec)}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			log.Fatalf("failed configuring registry: %s", err)
		}
	}
	return r
}

func (r *TypeRegistry) Register(specs ...*TypeSpec) error {
	for _, s := range specs {
		if _, exists := r.findType(s.Type); exists {
			return fmt.Errorf("type with T of %q already registered", s.Type)
		}
		if _, exists := r.findName(s.Name); exists {
			return fmt.Errorf("type with name of %q already registered", s.Type)
		}
		r.types = append(r.types, s)
		for _, alias := range s.Aliases {
			if _, exists := r.findType(alias); exists {
				return fmt.Errorf("type with T of %q already registered", s.Type)
			}
			r.aliases[alias] = s
		}
	}
	return nil
}

func (r *TypeRegistry) Specs() []*TypeSpec {
	return r.types[:]
}

func (r *TypeRegistry) PrintType(typ *ConcreteType) (string, error) {
	if r.format != nil {
		return r.format(typ)
	}
	return typ.Type, nil
}

func (r *TypeRegistry) FindSpecNamed(s string) (*TypeSpec, bool) {
	return r.findName(s)
}

func (r *TypeRegistry) findName(name string) (*TypeSpec, bool) {
	for _, current := range r.types {
		if current.Name == name {
			return current, true
		}
	}
	return nil, false
}

func (r *TypeRegistry) findType(t string) (*TypeSpec, bool) {
	for _, current := range r.types {
		if current.Type == t {
			return current, true
		}
	}
	return nil, false
}

func (r *TypeRegistry) Convert(typ Type) (*ConcreteType, error) {
	if ut, ok := typ.(*UnsupportedType); ok {
		return &ConcreteType{Type: ut.T}, nil
	}

	if r.spec == nil {
		return nil, fmt.Errorf("no spec function provided")
	}

	sp, err := r.spec(typ)
	if err != nil {
		return nil, err
	}

	if spec, ok := r.findType(sp.Type); ok {
		sp.Type = spec.Name
	} else if spec, ok := r.aliases[sp.Type]; ok {
		sp.Type = spec.Name
	} else {
		return nil, fmt.Errorf("unknown type %q", sp.Type)
	}
	return sp, nil
}

type TypeRegistryOption func(*TypeRegistry) error

func WithSpecs(specs ...*TypeSpec) TypeRegistryOption {
	return func(registry *TypeRegistry) error {
		if err := registry.Register(specs...); err != nil {
			return fmt.Errorf("failed registering types: %s", err)
		}
		return nil
	}
}

func WithSpecFunc(spec func(Type) (*ConcreteType, error)) TypeRegistryOption {
	return func(registry *TypeRegistry) error {
		registry.spec = spec
		return nil
	}
}

func WithParser(parser func(string) (Type, error)) TypeRegistryOption {
	return func(registry *TypeRegistry) error {
		registry.parser = parser
		return nil
	}
}

func WithFormatter(f func(*ConcreteType) (string, error)) TypeRegistryOption {
	return func(registry *TypeRegistry) error {
		registry.format = f
		return nil
	}
}
