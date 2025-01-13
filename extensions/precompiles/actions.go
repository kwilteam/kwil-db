// package actions allows custom actions to be registered with the
// engine.
package precompiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// Initializer is a function that creates a new instance of an extension.
// It is called:
// - Each time an extension is instantiated using `USE ... AS ...`
// - Once for every instantiated extension on node startup
// It should be used for reading values into memory, creating
// connections, and other setup that should only be done once per
// extension instance.
type Initializer func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]Value) (Instance, error)

// Instance is a named initialized instance of a precompile. It is
// returned from the precompile initialization, as specified by the
// Initializer. It will exist for the lifetime of the deployed
// dataset, and a single dataset can have multiple instances of the
// same precompile.
type Instance interface {
	// OnStart is called when the node starts, or when the extension is
	// first used. It is called right after Initialize, and before any
	// other methods are called.
	OnStart(ctx context.Context, app *common.App) error
	// OnUse is called when a `USE ... AS ...` statement is executed.
	// It is only called once per "USE" function, and is called after the
	// initializer.
	// It should be used for setting up state such as tables, indexes,
	// and other data structures that are part of the application state.
	OnUse(ctx *common.EngineContext, app *common.App) error
	// Methods returns the methods that are available on the instance.
	Methods() []*ExportedMethod
	// OnUnuse is called when a `UNUSE ...` statement is executed.
	OnUnuse(ctx *common.EngineContext, app *common.App) error
}

// ConcreteInstance is a concrete implementation of an extension instance.
type PrecompileExtension[T any] struct {
	// Initialize is the function that creates a new instance of the extension.
	Initialize func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]Value) (*T, error)
	// OnStart is called when the node starts, or when the extension is first used
	OnStart func(ctx context.Context, app *common.App, t *T) error
	// OnUse is called when a `USE ... AS ...` statement is executed
	OnUse func(ctx *common.EngineContext, app *common.App, t *T) error
	// Methods is a map of method names to method implementations.
	Methods []Method[T]
	// OnUnuse is called when a `UNUSE ...` statement is executed
	OnUnuse func(ctx *common.EngineContext, app *common.App, t *T) error
}

// Export exports the extension to a form that does not rely on generics, allowing the extension to be consumed by callers without forcing
// the callers to know the generic type.
func (p *PrecompileExtension[T]) export() Initializer {
	return func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]Value) (Instance, error) {
		var t *T
		if p.Initialize == nil {
			t = new(T)
		} else {
			t2, err := p.Initialize(ctx, service, db, alias, metadata)
			if err != nil {
				return nil, err
			}

			t = t2
		}

		methods := make([]*ExportedMethod, len(p.Methods))
		for i, method := range p.Methods {
			methods[i] = method.export(t)
		}

		return &ExportedExtension{
			onStart: func(ctx context.Context, app *common.App) error {
				if p.OnStart == nil {
					return nil
				}

				return p.OnStart(ctx, app, t)
			},
			onUse: func(ctx *common.EngineContext, app *common.App) error {
				if p.OnUse == nil {
					return nil
				}
				return p.OnUse(ctx, app, t)
			},
			methods: methods,
			onUnuse: func(ctx *common.EngineContext, app *common.App) error {
				if p.OnUnuse == nil {
					return nil
				}
				return p.OnUnuse(ctx, app, t)
			},
		}, nil
	}
}

type ExportedExtension struct {
	methods []*ExportedMethod
	onStart func(ctx context.Context, app *common.App) error
	onUse   func(ctx *common.EngineContext, app *common.App) error
	onUnuse func(ctx *common.EngineContext, app *common.App) error
}

func (e *ExportedExtension) OnStart(ctx context.Context, app *common.App) error {
	return e.onStart(ctx, app)
}

func (e *ExportedExtension) OnUse(ctx *common.EngineContext, app *common.App) error {
	return e.onUse(ctx, app)
}

func (e *ExportedExtension) Methods() []*ExportedMethod {
	return e.methods
}

func (e *ExportedExtension) OnUnuse(ctx *common.EngineContext, app *common.App) error {
	return e.onUnuse(ctx, app)
}

type Method[T any] struct {
	// Name is the name of the method.
	// It is case-insensitive, and should be unique within the extension.
	Name string
	// AccessModifiers is a list of access modifiers for the method.
	// It must have exactly one of PUBLIC, PRIVATE, or SYSTEM,
	// and can have any number of other modifiers.
	AccessModifiers []Modifier
	// Parameters is a list of parameters.
	// The engine will enforce that anything calling the method
	// provides the correct number of parameters, and in the correct
	// types.
	Parameters []*types.DataType
	// Returns specifies the return structure of the method.
	// If nil, the method does not return anything.
	Returns *MethodReturn
	// Handler is the function that is called when the method is invoked.
	Handler func(ctx *common.EngineContext, app *common.App, inputs []Value, resultFn func([]Value) error, t *T) error
}

func (m *Method[T]) verify() error {
	if strings.ToLower(m.Name) != m.Name {
		return fmt.Errorf("method name %s must be lowercase", m.Name)
	}

	if len(m.AccessModifiers) == 0 {
		return fmt.Errorf("method %s has no access modifiers", m.Name)
	}

	found := 0
	for _, mod := range m.AccessModifiers {
		if mod == PUBLIC || mod == PRIVATE || mod == SYSTEM {
			found++
		}
	}

	if found != 1 {
		return fmt.Errorf("method %s must have exactly one of PUBLIC, PRIVATE, or SYSTEM", m.Name)
	}

	if m.Returns != nil {
		if len(m.Returns.ColumnTypes) == 0 {
			return fmt.Errorf("method %s has no return types", m.Name)
		}

		if len(m.Returns.ColumnNames) != 0 && len(m.Returns.ColumnNames) != len(m.Returns.ColumnTypes) {
			return fmt.Errorf("method %s has %d return names, but %d return types", m.Name, len(m.Returns.ColumnNames), len(m.Returns.ColumnTypes))
		}
	}

	return nil
}

// MethodReturn specifies the return structure of a method.
type MethodReturn struct {
	// If true, then the method returns any number of rows.
	// If false, then the method returns exactly one row.
	ReturnsTable bool
	// ColumnTypes is a list of column types.
	// It is required. If the extension returns types that are
	// not matching the column types, the engine will return an error.
	ColumnTypes []*types.DataType
	// ColumnNames is a list of column names.
	// It is optional. If it is set, its length must be equal to the length
	// of the column types. If it is not set, the column names will be generated
	// based on their position in the column types.
	ColumnNames []string
}

// Modifier modifies the access to a procedure.
type Modifier string

const (
	// PUBLIC means that the action is public.
	PUBLIC Modifier = "PUBLIC"
	// PRIVATE means that the action is private.
	PRIVATE Modifier = "PRIVATE"
	// SYSTEM means that the action can only be called by the system.
	SYSTEM Modifier = "SYSTEM"
	// View means that an action does not modify the database.
	VIEW Modifier = "VIEW"

	// Owner requires that the caller is the owner of the database.
	OWNER Modifier = "OWNER"
)

type Modifiers []Modifier

func (m Modifiers) Has(mod Modifier) bool {
	for _, mod2 := range m {
		if mod2 == mod {
			return true
		}
	}
	return false
}

// export exports the method to a form that does not rely on generics, allowing the method to be consumed by callers without forcing
// the callers to know the generic type.
func (m *Method[T]) export(t *T) *ExportedMethod {
	return &ExportedMethod{
		Name:            m.Name,
		AccessModifiers: m.AccessModifiers,
		Parameters:      m.Parameters,
		Returns:         m.Returns,
		Call: func(ctx *common.EngineContext, app *common.App, inputs []Value, resultFn func([]Value) error) error {
			return m.Handler(ctx, app, inputs, resultFn, t)
		},
	}
}

type ExportedMethod struct {
	Name            string
	AccessModifiers []Modifier
	Parameters      []*types.DataType
	Returns         *MethodReturn
	Call            func(ctx *common.EngineContext, app *common.App, inputs []Value, resultFn func([]Value) error) error
}

var registeredPrecompiles = make(map[string]Initializer)

func RegisteredPrecompiles() map[string]Initializer {
	return registeredPrecompiles
}

// RegisterPrecompile registers a precompile extension with the
// engine.
func RegisterPrecompile[T any](name string, ext PrecompileExtension[T]) error {
	name = strings.ToLower(name)
	if _, ok := registeredPrecompiles[name]; ok {
		return fmt.Errorf("precompile of same name already registered:%s ", name)
	}

	methodNames := make(map[string]struct{})
	for _, method := range ext.Methods {
		err := method.verify()
		if err != nil {
			return fmt.Errorf("method %s: %w", method.Name, err)
		}

		if _, ok := methodNames[method.Name]; ok {
			return fmt.Errorf("duplicate method %s", method.Name)
		}

		methodNames[method.Name] = struct{}{}
	}

	registeredPrecompiles[name] = ext.export()
	return nil
}
