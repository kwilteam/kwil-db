package interpreter

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// initializeExtension initializes an extension.
func initializeExtension(ctx context.Context, svc *common.Service, db sql.DB, i precompiles.Initializer, alias string, metadata map[string]Value) (*namespace, error) {
	convertedMetadata := make(map[string]any)
	for k, v := range metadata {
		convertedMetadata[k] = v.RawValue()
	}

	inst, err := i(ctx, svc, db, alias, convertedMetadata)
	if err != nil {
		return nil, err
	}

	// we construct a map of executables
	executables := copyBuiltinExecutables()
	methods := make(map[string]*executable)
	for _, method := range inst.Methods() {
		lowerName := strings.ToLower(method.Name)

		exec := &executable{
			Name: lowerName,
			Func: func(exec *executionContext, args []Value, fn resultFunc) error {
				if err := exec.canExecute(alias, lowerName, method.AccessModifiers); err != nil {
					return err
				}

				argVals := make([]any, len(args))
				for i, arg := range args {
					argVals[i] = arg.RawValue()
				}

				exec2 := exec.subscope(alias)

				return method.Call(exec2.engineCtx, exec2.app(), argVals, func(a []any) error {
					resultVals := make([]Value, len(a))
					for i, result := range a {
						var err error
						resultVals[i], err = NewValue(result)
						if err != nil {
							return err
						}
					}

					if len(method.ReturnColumns) != 0 && len(method.ReturnColumns) != len(resultVals) {
						return fmt.Errorf("method %s returned %d values, but expected %d", method.Name, len(resultVals), len(method.ReturnColumns))
					}

					return fn(&row{
						columns: method.ReturnColumns, // it is ok if this is nil
						Values:  resultVals,
					})
				})
			},
			Type: executableTypePrecompile,
		}

		executables[lowerName] = exec
		methods[lowerName] = exec
	}

	return &namespace{
		availableFunctions: executables,
		tables:             make(map[string]*engine.Table),
		onDeploy: func(ctx *executionContext) error {
			return inst.OnUse(ctx.engineCtx, ctx.app())
		},
		onUndeploy: func(ctx *executionContext) error {
			return inst.OnUnuse(ctx.engineCtx, ctx.app())
		},
		namespaceType: namespaceTypeExtension,
		methods:       methods,
	}, nil
}
