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
func initializeExtension(ctx context.Context, svc *common.Service, db sql.DB, i precompiles.Initializer, alias string,
	metadata map[string]precompiles.Value) (*namespace, precompiles.Instance, error) {

	inst, err := i(ctx, svc, db, alias, metadata)
	if err != nil {
		return nil, nil, err
	}

	// we construct a map of executables
	executables := copyBuiltinExecutables()
	methods := make(map[string]*executable)
	for _, method := range inst.Methods() {
		lowerName := strings.ToLower(method.Name)

		exec := &executable{
			Name: lowerName,
			Func: func(exec *executionContext, args []precompiles.Value, fn resultFunc) error {
				if err := exec.canExecute(alias, lowerName, method.AccessModifiers); err != nil {
					return err
				}

				argVals := make([]any, len(args))
				for i, arg := range args {
					argVals[i] = arg.RawValue()
				}

				exec2 := exec.subscope(alias)

				return method.Call(exec2.engineCtx, exec2.app(), args, func(a []precompiles.Value) error {

					var colNames []string
					if method.Returns != nil {
						if len(method.Returns.ColumnTypes) != len(a) {
							return fmt.Errorf("method %s returned %d values, but expected %d", method.Name, len(a), len(method.Returns.ColumnTypes))
						}

						for i, result := range a {
							if !result.Type().Equals(method.Returns.ColumnTypes[i]) {
								return fmt.Errorf("method %s returned a value of type %s, but expected %s", method.Name, result.Type(), method.Returns.ColumnTypes[i])
							}
						}

						if len(method.Returns.ColumnNames) > 0 {
							colNames = method.Returns.ColumnNames
						}
					}

					return fn(&row{
						columns: colNames, // it is ok if this is nil
						Values:  a,
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
	}, inst, nil
}
