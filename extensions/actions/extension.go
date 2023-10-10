package extensions

import (
	"context"
	"fmt"
)

// Local Extension
type Extension struct {
	// Extension name
	name string
	// Supported methods by the extension
	methods map[string]MethodFunc
	// Initializer that initializes the extension
	initializeFunc InitializeFunc
}

func (e *Extension) Name() string {
	return e.name
}

func (e *Extension) Execute(ctx context.Context, metadata map[string]string, method string, args ...any) ([]any, error) {
	var encodedArgs []*ScalarValue
	for _, arg := range args {
		scalarVal, err := NewScalarValue(arg)
		if err != nil {
			return nil, fmt.Errorf("error encoding argument: %s", err.Error())
		}

		encodedArgs = append(encodedArgs, scalarVal)
	}

	methodFn, ok := e.methods[method]
	if !ok {
		return nil, fmt.Errorf("method %s not found", method)
	}

	execCtx := &ExecutionContext{
		Ctx:      ctx,
		Metadata: metadata,
	}
	results, err := methodFn(execCtx, encodedArgs...)
	if err != nil {
		return nil, err
	}

	var outputs []any
	for _, result := range results {
		outputs = append(outputs, result.Value)
	}
	return outputs, nil
}

func (e *Extension) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	return e.initializeFunc(ctx, metadata)
}
