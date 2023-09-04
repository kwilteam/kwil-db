package execution

import (
	"context"
)

const (
	callerVarName        = "@caller"
	callerAddressVarName = "@caller_address"

	actionVarName = "@action"
	defaultAction = "_no_action_"

	datasetVarName = "@dataset"
	datasetDefault = "x00000000000000000000000000000000000000000000000000000000"
)

// executionContext is the context for executing a block of code.
// It should be created with newExecutionContext.
type executionContext struct {
	ctx           context.Context
	caller        User
	action        string
	datasetID     string
	lastDmlResult []map[string]any

	// mustBeNonMutative is true if the execution context must be non-mutative.
	// if true, execution of mutative statements will fail.
	mustBeNonMutative bool
}

func (ec *executionContext) contextualVariables() map[string]any {
	return map[string]any{
		callerVarName:        ec.caller.Bytes(),
		callerAddressVarName: ec.caller.Address(),
		actionVarName:        ec.action,
		datasetVarName:       ec.datasetID,
	}
}

func newExecutionContext(ctx context.Context, action string, opts ...ExecutionOpt) *executionContext {
	ec := &executionContext{
		ctx:       ctx,
		caller:    &noCaller{},
		action:    action,
		datasetID: datasetDefault,
	}

	for _, opt := range opts {
		opt(ec)
	}

	return ec
}

type ExecutionOpt func(*executionContext)

func WithCaller(caller User) ExecutionOpt {
	return func(ec *executionContext) {
		ec.caller = caller
	}
}

func WithDatasetID(dataset string) ExecutionOpt {
	return func(ec *executionContext) {
		ec.datasetID = dataset
	}
}

func NonMutative() ExecutionOpt {
	return func(ec *executionContext) {
		ec.mustBeNonMutative = true
	}
}

// a procedureContext is the context for executing a procedure.
// it contains the executionContext, as well as the values scoped to the procedure.
type procedureContext struct {
	*executionContext
	values map[string]any
}
