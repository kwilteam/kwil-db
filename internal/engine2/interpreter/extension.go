package interpreter

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/engine2"
)

// Initializer is a function that creates a new instance of an extension.
// It is called:
// - Each time an extension is instantiated using `USE ... AS ...`
// - Once for every instantiated extension on node startup
// It should be used for reading values into memory, creating
// connections, and other setup that should only be done once per
// extension instance.
type Initializer func(ctx context.Context, service *common.Service, db sql.DB, metadata map[string]Value) (Instance, error)

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
	OnUse(ctx *common.TxContext, app *common.App, metadata map[string]Value) error
	// Call executes the requested method of the precompile. It is up
	// to the instance implementation to determine if a method is
	// valid, and to subsequently decode the arguments. The arguments
	// passed in as args, as well as returned, are scalar values.
	Call(ctx *common.TxContext, app *common.App, method string, inputs []Value, resultFn func([]Value)) error
	// OnUnuse is called when a `UNUSE ...` statement is executed.
	OnUnuse(ctx *common.TxContext, app *common.App) error
}

// ConcreteInstance is a concrete implementation of an extension instance.
type PrecompileExtension[T any] struct {
	// Initialize is the function that creates a new instance of the extension.
	Initialize func(ctx context.Context, service *common.Service, db sql.DB, metadata map[string]Value) (*T, error)
	// OnUse is called when a `USE ... AS ...` statement is executed
	OnUse func(ctx *common.TxContext, app *common.App, metadata map[string]Value, t *T) error
	// Methods is a map of method names to method implementations.
	Methods []*Method[T]
	// OnUnuse is called when a `UNUSE ...` statement is executed
	OnUnuse func(ctx *common.TxContext, app *common.App, t *T) error
}

type Method[T any] struct {
	// Name is the name of the method.
	// It is case-insensitive, and should be unique within the extension.
	Name string
	// AccessModifiers is a list of access modifiers for the method.
	// It must have exactly one of PUBLIC, PRIVATE, or SYSTEM,
	// and can have any number of other modifiers.
	AccessModifiers []Modifier
	// Call is the function that is called when the method is invoked.
	Call func(ctx *common.TxContext, app *common.App, inputs []Value, resultFn func([]Value) error, t *T) error
}

func initializeExtension[T any](exec *executionContext, p *PrecompileExtension[T], alias string, metadata map[string]Value) (*namespace, error) {
	var t *T
	if p.Initialize != nil {
		t2, err := p.Initialize(exec.txCtx.Ctx, exec.interpreter.service, exec.db, metadata)
		if err != nil {
			return nil, err
		}

		t = t2
	} else {
		t = new(T)
	}

	executables, err := makeExtensionExecutables(alias, p.Methods, t)
	if err != nil {
		return nil, err
	}

	// finally, we add two special methods: use and unuse
	executables[useMethod] = &executable{
		Name: useMethod,
		Func: func(exec *executionContext, args []Value, returnFn func([]Value) error) error {
			if p.OnUse == nil {
				return nil
			}
			return p.OnUse(exec.txCtx, exec.app(), metadata, t)
		},
		Type: executableTypePrecompile,
	}

	executables[unuseMethod] = &executable{
		Name: unuseMethod,
		Func: func(exec *executionContext, args []Value, returnFn func([]Value) error) error {
			if p.OnUnuse == nil {
				return nil
			}
			return p.OnUnuse(exec.txCtx, exec.app(), t)
		},
		Type: executableTypePrecompile,
	}

	return &namespace{
		availableFunctions: executables,
		tables:             make(map[string]*engine2.Table),
		namespaceType:      namespaceTypeExtension,
	}, nil
}

const (
	useMethod   = "use"
	unuseMethod = "unuse"
)

func executeOnUse(exec *executionContext, ns *namespace) error {
	useExec, ok := ns.availableFunctions[useMethod]
	if !ok {
		return fmt.Errorf("missing use method")
	}

	return useExec.Func(exec, nil, nil)
}

func executeOnUnuse(exec *executionContext, ns *namespace) error {
	unuseExec, ok := ns.availableFunctions[unuseMethod]
	if !ok {
		return fmt.Errorf("missing unuse method")
	}

	return unuseExec.Func(exec, nil, nil)
}

func makeExtensionExecutables[T any](namespace string, methods []*Method[T], t *T) (map[string]*executable, error) {
	execs := copyBuiltinExecutables()
	methNames := make(map[string]struct{})

	for _, method := range methods {
		methName := strings.ToLower(method.Name)
		if _, ok := methNames[methName]; ok {
			return nil, fmt.Errorf("duplicate method name %s", methName)
		}

		methNames[methName] = struct{}{}

		if methName == useMethod || methName == unuseMethod {
			return nil, fmt.Errorf("method name %s is reserved", methName)
		}

		mods := slices.Clone(method.AccessModifiers) // avoid loop variable capture
		execs[methName] = &executable{
			Name: methName,
			Func: func(exec *executionContext, args []Value, returnFn func([]Value) error) error {
				err := exec.canExecute(namespace, methName, mods)
				if err != nil {
					return err
				}

				// TODO: add the missing parts
				return method.Call(nil, nil, args, returnFn, t)
			},
			Type: executableTypePrecompile,
		}
	}

	return execs, nil
}
