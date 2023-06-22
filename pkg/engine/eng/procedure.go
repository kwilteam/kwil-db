package eng

import "fmt"

// a procedure is a collection of operations that can be executed as a single unit
// it is atomic, and has local variables
type Procedure struct {
	Name       string                  `json:"name"`
	Parameters []string                `json:"args"`
	Scoping    ProcedureScoping        `json:"scoping"`
	Body       []*InstructionExecution `json:"body"`
}

func (p *Procedure) evaluate(ctx *executionContext, eng *Engine, ins []*InstructionExecution, args ...any) error {
	if len(args) != len(p.Parameters) {
		return fmt.Errorf("%w: procedure '%s' requires %d arguments, got %d", ErrIncorrectNumArgs, p.Name, len(p.Parameters), len(args))
	}

	vars := ctx.contextualVariables()
	for i, arg := range args {
		vars[p.Parameters[i]] = arg
	}

	return evaluateInstructions(ctx, eng, ins, vars)
}

func (a *Procedure) checkAccessControl(opts *executionContext) error {
	if a.Scoping == ProcedureScopingPublic {
		return nil
	}

	if opts == nil {
		return fmt.Errorf("failed to execute private action '%s': could not authenticate caller", a.Name)
	}

	if opts.action == a.Name {
		return fmt.Errorf("%w: private action cannot be invoked directly", ErrScopingViolation)
	}

	return nil
}

type ProcedureScoping uint8

func (p ProcedureScoping) Clean() error {
	if p != ProcedureScopingPublic && p != ProcedureScopingPrivate {
		return fmt.Errorf("invalid procedure scoping '%d'", p)
	}

	return nil
}

const (
	ProcedureScopingPublic ProcedureScoping = iota
	ProcedureScopingPrivate
)
