package actions

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	actparser "github.com/kwilteam/kwil-db/parse/actions/parser"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/clean"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/parameters"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// AnalyzeOpts are options for analyzing actions.
type AnalyzeOpts struct {
	// PGSchemaName is the name of the Postgres schema.
	// If not set, it will default to the passed schema dbid.
	PGSchemaName string
	// SchemaInfo contains position information for the schema.
	// If not nil, it will use the position information to modify the error messages.
	SchemaInfo *parseTypes.SchemaInfo
}

// AnalyzeActions analyzes the actions in a schema.
// It will perform validation checks on statements, such as ensuring that
// all view actions do not modify state. It will make all sql statements deterministic.
func AnalyzeActions(schema *types.Schema, opts *AnalyzeOpts) ([]*AnalyzedAction, parseTypes.ParseErrors, error) {
	if opts == nil {
		opts = &AnalyzeOpts{}
	}

	if opts.PGSchemaName == "" {
		opts.PGSchemaName = schema.DBID()
	}

	parseErrs := parseTypes.ParseErrors{}

	analyzed := make([]*AnalyzedAction, len(schema.Actions))
	for i, action := range schema.Actions {
		// if schema info is provided, we will use it to modify the error messages.
		errorListener := parseTypes.NewErrorListener()
		startingLine := 1
		startingCol := 1
		if opts.SchemaInfo != nil {
			actionPos, ok := opts.SchemaInfo.Blocks[action.Name]
			if !ok {
				// should never happen, this would be a bug in our code
				return nil, nil, fmt.Errorf("could not find position for action %s", action.Name)
			}

			startingLine = actionPos.StartLine
			startingCol = actionPos.StartCol
		}

		errorListener = errorListener.Child("action", startingLine, startingCol)
		stmts, err := actparser.Parse(action.Body, errorListener)
		if err != nil {
			return nil, nil, err
		}

		analyzedStmts := make([]AnalyzedStatement, len(stmts))

		for j, stmt := range stmts {
			analyzedStmt, err := convertStatement(stmt, schema, opts.PGSchemaName, errorListener)
			if err != nil {
				return nil, nil, err
			}
			analyzedStmts[j] = analyzedStmt
		}

		analyzed[i] = &AnalyzedAction{
			Name:       strings.ToLower(action.Name),
			Public:     action.Public,
			IsView:     action.IsView(),
			OwnerOnly:  action.IsOwnerOnly(),
			Parameters: action.Parameters,
			Statements: analyzedStmts,
		}

		parseErrs.Add(errorListener.Errors()...)
	}

	return analyzed, parseErrs, nil
}

// convertStatement converts a statement from the actparser AST to an AnalyzedStatement.
func convertStatement(stmt actparser.ActionStmt, schema *types.Schema, pgSchemaName string, errLis parseTypes.NativeErrorListener) (AnalyzedStatement, error) {
	switch stmt := stmt.(type) {
	case *actparser.ExtensionCallStmt:
		recs := make([]string, len(stmt.Receivers))
		for i, rec := range stmt.Receivers {
			recs[i] = strings.ToLower(rec)
		}

		params, err := prepareInlineExpressions(stmt.Args, schema, errLis)
		if err != nil {
			return nil, err
		}

		return &ExtensionCall{
			Extension: strings.ToLower(stmt.Extension),
			Method:    strings.ToLower(stmt.Method),
			Params:    params,
			Receivers: recs,
		}, nil
	case *actparser.ActionCallStmt:
		// receivers and targets must be empty for actions.
		// I am not really sure why these were ever added, as they
		// were never supported
		if stmt.Database != "" {
			return nil, fmt.Errorf("cannot call actions in other databases")
		}
		if len(stmt.Receivers) > 0 {
			return nil, fmt.Errorf("actions cannot specify return values")
		}

		params, err := prepareInlineExpressions(stmt.Args, schema, errLis)
		if err != nil {
			return nil, err
		}

		return &ActionCall{
			Action: strings.ToLower(stmt.Method),
			Params: params,
		}, nil
	case *actparser.DMLStmt:
		child := errLis.Child("action-sql", stmt.Node.StartLine, stmt.Node.StartCol)
		// make the statement deterministic.
		generated, err := sqlanalyzer.ApplyRules(stmt.Statement, sqlanalyzer.AllRules, schema, pgSchemaName, child)
		if err != nil {
			return nil, err
		}
		if child.Err() != nil {
			errLis.Add(child.Errors()...)
		}

		return &SQLStatement{
			Statement:      generated.Statement,
			Mutative:       generated.Mutative,
			ParameterOrder: generated.ParameterOrder,
		}, nil
	}

	return nil, fmt.Errorf("unknown statement type: %T", stmt)
}

// AnalyzedAction contains the results of analyzing an action.
type AnalyzedAction struct {
	// Name is the name of the action.
	Name string
	// Public is true if the action is public.
	Public bool
	// IsView is true if the action is a view.
	IsView bool
	// OwnerOnly is true if the action is owner-only.
	OwnerOnly bool
	// Parameters is a list of parameters for the action.
	Parameters []string
	// Statements are the statements in the action.
	Statements []AnalyzedStatement
}

// AnalyzedStatement is an interface for analyzed statements.
type AnalyzedStatement interface {
	analyzedStmt()
}

// there are exactly three types of analyzed statements:
// - ExtensionCall: a statement that calls an extension
// - ActionCall: a statement that calls an action
// - SQLStatement: a statement that contains SQL

// ExtensionCall is an analyzed statement that calls an action or extension.
type ExtensionCall struct {
	// Extension is the name of the extension alias.
	Extension string
	// Method is the name of the method being called.
	Method string
	// Params are the parameters to the method.
	Params []*InlineExpression
	// Receivers are the receivers of the method.
	Receivers []string
}

func (c *ExtensionCall) analyzedStmt() {}

// ActionCall is an analyzed statement that calls an action.
type ActionCall struct {
	// Action is the name of the action being called.
	Action string
	// Params are the parameters to the action.
	Params []*InlineExpression
}

func (c *ActionCall) analyzedStmt() {}

// SQLStatement is an analyzed statement that contains SQL.
type SQLStatement struct {
	// Statement is the Statement statement that should be executed.
	// It is deterministic.
	Statement string
	// Mutative is true if the statement mutates state.
	Mutative bool
	// ParameterOrder is a list of the parameters in the order they appear in the statement.
	// This is set if the ReplaceNamedParameters flag is set.
	// For example, if the statement is "SELECT * FROM table WHERE id = $id AND name = @caller",
	// then the parameter order would be ["$id", "@caller"]
	ParameterOrder []string
}

func (s *SQLStatement) analyzedStmt() {}

// prepareInlineExpressions prepares inline expressions for analysis.
// It takes the expressions from the syntax tree, as well as the procedures
// that exist in the schema, which is necessary for validating the expressions.
func prepareInlineExpressions(exprs []tree.Expression, schema *types.Schema, errLis parseTypes.NativeErrorListener) ([]*InlineExpression, error) {
	prepared := make([]*InlineExpression, len(exprs))
	for i, expr := range exprs {
		// this is copied over from an old place in the engine.
		switch e := expr.(type) {
		case *tree.ExpressionBindParameter:
			// This could be a special one that returns an evaluatable that
			// ignores the passed ResultSetFunc since the value is
		case *tree.ExpressionTextLiteral, *tree.ExpressionIntLiteral, *tree.ExpressionBooleanLiteral,
			*tree.ExpressionNullLiteral, *tree.ExpressionBlobLiteral, *tree.ExpressionUint256Literal,
			*tree.ExpressionDecimalLiteral, *tree.ExpressionUnary, *tree.ExpressionBinaryComparison,
			*tree.ExpressionFunction, *tree.ExpressionArithmetic:
			// Acceptable expression type.
		default:
			return nil, fmt.Errorf("unsupported expression type: %T", e)
		}

		// clean expression, since it is submitted by the user
		err := expr.Walk(clean.NewStatementCleaner(schema, errLis))
		if err != nil {
			return nil, err
		}

		// The schema walker is not necessary for inline expressions, since
		// we do not support table references in inline expressions.
		walker := sqlanalyzer.NewWalkerRecoverer(expr)
		paramVisitor := parameters.NewParametersWalker()
		err = walker.Walk(paramVisitor)
		if err != nil {
			return nil, fmt.Errorf("error replacing parameters: %w", err)
		}

		// SELECT expr;  -- prepare new value in receivers for subsequent
		// statements This query needs to be run in "simple" execution mode
		// rather than "extended" execution mode, which asks the database for
		// OID (placeholder types) that it can't know since there's no FOR table.
		selectTree := &tree.SelectStmt{
			Stmt: &tree.SelectCore{
				SimpleSelects: []*tree.SimpleSelect{
					{
						SelectType: tree.SelectTypeAll,
						Columns: []tree.ResultColumn{
							&tree.ResultColumnExpression{
								Expression: expr,
							},
						},
					},
				},
			},
		}

		stmt, err := tree.SafeToSQL(selectTree)
		if err != nil {
			return nil, err
		}

		prepared[i] = &InlineExpression{
			Statement:     stmt,
			OrderedParams: paramVisitor.OrderedParameters,
		}
	}
	return prepared, nil
}

// InlineExpression is an expression that is inlined in an action or procedure call.
// For example, this can be "extension.call($id+1)"
type InlineExpression struct {
	// Statement is the sql statement that is inlined.
	Statement string
	// OrderedParams is the order of the parameters in the statement.
	OrderedParams []string
}
