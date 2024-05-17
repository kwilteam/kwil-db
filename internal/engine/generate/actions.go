package generate

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

// this package handles generating code for actions.

// GeneratedActionStmt is an interface for analyzed statements.
type GeneratedActionStmt interface {
	generatedAction()
}

// there are exactly three types of analyzed statements:
// - ActionExtensionCall: a statement that calls an extension
// - ActionCall: a statement that calls an action
// - ActionSQL: a statement that contains SQL

// ActionExtensionCall is an analyzed statement that calls an action or extension.
type ActionExtensionCall struct {
	// Extension is the name of the extension alias.
	Extension string
	// Method is the name of the method being called.
	Method string
	// Params are the parameters to the method.
	Params []*InlineExpression
	// Receivers are the receivers of the method.
	Receivers []string
}

func (c *ActionExtensionCall) generatedAction() {}

// ActionCall is an analyzed statement that calls an action.
type ActionCall struct {
	// Action is the name of the action being called.
	Action string
	// Params are the parameters to the action.
	Params []*InlineExpression
}

func (c *ActionCall) generatedAction() {}

// ActionSQL is an analyzed statement that contains SQL.
type ActionSQL struct {
	// Statement is the Statement statement that should be executed.
	// It is deterministic.
	Statement string
	// ParameterOrder is a list of the parameters in the order they appear in the statement.
	// This is set if the ReplaceNamedParameters flag is set.
	// For example, if the statement is "SELECT * FROM table WHERE id = $id AND name = @caller",
	// then the parameter order would be ["$id", "@caller"]
	ParameterOrder []string
}

func (s *ActionSQL) generatedAction() {}

// InlineExpression is an expression that is inlined in an action or procedure call.
// For example, this can be "extension.call($id+1)"
type InlineExpression struct {
	// Statement is the sql statement that is inlined.
	Statement string
	// OrderedParams is the order of the parameters in the statement.
	OrderedParams []string
}

// GenerateActionBody generates the body of an action.
// If the action is a VIEW and contains mutative SQL, it will return an error.
func GenerateActionBody(action *types.Action, schema *types.Schema) (stmts []GeneratedActionStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", r)
			}
		}

		// add action name to error
		if err != nil {
			err = fmt.Errorf("action %s: %w", action.Name, err)
		}
	}()

	res, err := parse.ParseAction(action, schema)
	if err != nil {
		return nil, err
	}

	// syntax errors, as well as mutative SQL, will be thrown here.
	if res.ParseErrs.Err() != nil {
		return nil, res.ParseErrs.Err()
	}

	g := &actionGenerator{}
	for _, stmt := range res.AST {
		stmt.Accept(g)
	}

	return g.actions, nil
}

// actionGenerator is a struct that generates code for actions.
// it totally relies on the SQL generator, except for the action-specific
// visits and Variables, since it needs to rewrite the variables to be numbered.
type actionGenerator struct {
	sqlParamRewriter
	// actions is the order of all actions that are generated.
	actions []GeneratedActionStmt
	// paramOrder is the name of the parameters in the order they appear in the action.
	// Since the actionGenerator rewrites actions from named parameters to numbered parameters,
	// the order of the named parameters is stored here.
	paramOrder []string
}

func (a *actionGenerator) VisitActionStmtSQL(p0 *parse.ActionStmtSQL) any {
	a.paramOrder = nil
	stmt := p0.SQL.Accept(a).(string)

	a.actions = append(a.actions, &ActionSQL{
		Statement:      stmt + ";",
		ParameterOrder: a.paramOrder,
	})

	return nil
}

func (a *actionGenerator) VisitExtensionCallStmt(p0 *parse.ActionStmtExtensionCall) any {
	inlines := make([]*InlineExpression, len(p0.Args))
	for i, arg := range p0.Args {
		inlines[i] = a.createInline(arg)
	}

	a.actions = append(a.actions, &ActionExtensionCall{
		Receivers: p0.Receivers,
		Extension: p0.Extension,
		Method:    p0.Method,
		Params:    inlines,
	})

	return nil
}

func (a *actionGenerator) VisitActionCallStmt(p0 *parse.ActionStmtActionCall) any {
	inlines := make([]*InlineExpression, len(p0.Args))
	for i, arg := range p0.Args {
		inlines[i] = a.createInline(arg)
	}

	a.actions = append(a.actions, &ActionCall{
		Action: p0.Action,
		Params: inlines,
	})

	return nil
}

// createInline creates an inline from an expression
func (a *actionGenerator) createInline(p0 parse.Expression) *InlineExpression {
	a.paramOrder = nil

	str := p0.Accept(a).(string)

	params := make([]string, len(a.paramOrder))
	copy(params, a.paramOrder)

	return &InlineExpression{
		Statement:     "SELECT " + str + ";",
		OrderedParams: params,
	}
}
