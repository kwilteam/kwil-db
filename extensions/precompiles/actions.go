// package actions allows custom actions to be registered with the
// engine.
package precompiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// Initializer is a function that creates a new instance of an extension.
// It is called:
// - Each time an extension is instantiated using `USE ... AS ...`
// - Once for every instantiated extension on node startup
// It should be used for reading values into memory, creating
// connections, and other setup that should only be done once per
// extension instance.
type Initializer func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (Instance, error)

// Instance is a named initialized instance of a precompile. It is
// returned from the precompile initialization, as specified by the
// Initializer. It will exist for the lifetime of the deployed
// dataset, and a single dataset can have multiple instances of the
// same precompile.
type Instance interface {
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
	Initialize func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (*T, error)
	// OnUse is called when a `USE ... AS ...` statement is executed
	OnUse func(ctx *common.EngineContext, app *common.App, t *T) error
	// Methods is a map of method names to method implementations.
	Methods []Method[T]
	// OnUnuse is called when a `UNUSE ...` statement is executed
	OnUnuse func(ctx *common.EngineContext, app *common.App, t *T) error
}

// Export exports the extension to a form that does not rely on generics, allowing the extension to be consumed by callers without forcing
// the callers to know the generic type.
func (p *PrecompileExtension[T]) Export() Initializer {
	return func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (Instance, error) {
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
	onUse   func(ctx *common.EngineContext, app *common.App) error
	onUnuse func(ctx *common.EngineContext, app *common.App) error
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
	// Handler is the function that is called when the method is invoked.
	Handler func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, t *T) error
	// ReturnColumns is a list of the returned column names. It is optional. If it is set, its length must be
	// equal to the length of the returned values passed to the resultFn. If it is not set, the returned column
	// names will be generated based on their position in the returned values.
	ReturnColumns []string
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
		Call: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
			return m.Handler(ctx, app, inputs, resultFn, t)
		},
	}
}

type ExportedMethod struct {
	Name            string
	AccessModifiers []Modifier
	Call            func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error
	// ReturnColumns is a list of the returned column names. It is optional. If it is set, its length must be
	// equal to the length of the returned values passed to the resultFn.
	ReturnColumns []string
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
		lowerName := strings.ToLower(method.Name)
		if _, ok := methodNames[lowerName]; ok {
			return fmt.Errorf("duplicate method %s", lowerName)
		}

		methodNames[lowerName] = struct{}{}

		if len(method.AccessModifiers) == 0 {
			return fmt.Errorf("method %s has no access modifiers", method.Name)
		}

		found := 0
		for _, mod := range method.AccessModifiers {
			if mod == PUBLIC || mod == PRIVATE || mod == SYSTEM {
				found++
			}
		}

		if found != 1 {
			return fmt.Errorf("method %s must have exactly one of PUBLIC, PRIVATE, or SYSTEM", method.Name)
		}
	}

	registeredPrecompiles[name] = ext.Export()
	return nil
}
