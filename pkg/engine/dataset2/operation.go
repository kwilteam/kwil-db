package dataset2

import "fmt"

// an operation is a function that takes a context and a slice of arguments and returns a slice of results and an error
// things like dml statements and extension methods are operations
type operation interface {
	// evaluate evaluates the operation.
	evaluate(ctx *procedureContext, args ...any) (map[string]any, error)

	// requiredVariables returns the variables required by the operation.
	requiredVariables() []string

	close() error
}

type dmlStatement struct {
	stmt PreparedStatement
}

func (s *dmlStatement) evaluate(ctx *procedureContext, args ...any) (map[string]any, error) {
	return nil, nil
}

func (s *dmlStatement) requiredVariables() []string {
	// sql statements only use variables from the values map
	return []string{}
}

func (s *dmlStatement) close() error {
	return s.stmt.Close()
}

type extensionMethod struct {
	extensions   InitializedExtension
	method       string
	requiredArgs []string
	returns      []string
}

func (m *extensionMethod) evaluate(ctx *procedureContext, args ...any) (map[string]any, error) {
	if len(args) != len(m.requiredArgs) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(m.requiredArgs), len(args))
	}

	return nil, nil
}

func (m *extensionMethod) requiredVariables() []string {
	return m.requiredArgs
}

func (m *extensionMethod) close() error {
	return nil
}

type procedureExecution struct {
	procedure *StoredProcedure
	args      []string
}

func (e *procedureExecution) evaluate(ctx *procedureContext, args ...any) (map[string]any, error) {
	if len(args) != len(e.args) {
		return nil, fmt.Errorf("procedure '%s' expects %d arguments, got %d", e.procedure.Name, len(e.args), len(args))
	}

	err := e.procedure.Execute(ctx.executionContext, args)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (e *procedureExecution) requiredVariables() []string {
	return e.args
}

func (e *procedureExecution) close() error {
	return nil
}
