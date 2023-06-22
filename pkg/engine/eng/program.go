package eng

import "fmt"

type InstructionExecution struct {
	Instruction Instruction `json:"instruction"`
	Args        []any       `json:"args"`
}

func (ie *InstructionExecution) evaluate(ctx *procedureContext, eng *Engine) error {
	return ie.Instruction.evaluate(ctx, eng, ie.Args...)
}

// evaluateInstructions takes a list of instructions and executes them in order.
func evaluateInstructions(ctx *executionContext, eng *Engine, ins []*InstructionExecution, values map[string]any) error {
	if values == nil {
		values = make(map[string]any)
	}

	procedureCtx := &procedureContext{
		executionContext: ctx,
		values:           values,
	}

	sp, err := eng.db.Savepoint()
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	for _, instruction := range ins {
		err = instruction.evaluate(procedureCtx, eng)
		if err != nil {
			return fmt.Errorf("failed to evaluate instruction: %w", err)
		}
	}

	err = sp.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit savepoint: %w", err)
	}

	return nil
}
