package execution

import (
	"context"
	"fmt"
	"strings"
)

func (e *Engine) ExecuteProcedure(ctx context.Context, name string, args []any, opts ...ExecutionOpt) ([]map[string]any, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	proc, ok := e.procedures[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("procedure %s not found", name)
	}

	execCtx := newExecutionContext(ctx, name, opts...)

	err := proc.checkAccessControl(execCtx)
	if err != nil {
		return nil, fmt.Errorf("access control failed: %w", err)
	}

	err = proc.evaluate(execCtx, e, proc.Body, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute procedure: %w", err)
	}

	return execCtx.lastDmlResult, nil
}

const (
	loadCmdName = "_load"
)

func (e *Engine) executeLoad(ctx context.Context) error {
	if len(e.loadCommand) == 0 {
		return nil
	}

	execCtx := newExecutionContext(ctx, loadCmdName)

	return evaluateInstructions(execCtx, e, e.loadCommand, nil)
}
