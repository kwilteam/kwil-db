package dataset2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

const (
	defaultCallerAddress = "0x0000000000000000000000000000000000000000"
	callerVarName        = "@caller"

	actionVarName = "@action"
	defaultAction = "_no_action_"

	datasetVarName = "@dataset"
	datasetDefault = "x00000000000000000000000000000000000000000000000000000000"
)

// executionContext is the context for executing a block of code.
// It should be created with newExecutionContext.
type executionContext struct {
	ctx           context.Context
	caller        string
	action        string
	dataset       string
	lastDmlResult dto.Result
}

func (ec *executionContext) contextualVariables() map[string]any {
	return map[string]any{
		callerVarName:  ec.caller,
		actionVarName:  ec.action,
		datasetVarName: ec.dataset,
	}
}

func newExecutionContext(ctx context.Context, action string, opts ...ExecutionOpt) *executionContext {
	ec := &executionContext{
		ctx:     ctx,
		caller:  defaultCallerAddress,
		action:  defaultAction,
		dataset: datasetDefault,
	}

	for _, opt := range opts {
		opt(ec)
	}

	return ec
}
