/*
Package evaluater provides a way for preparing and evluating arbitrary expressions.
This is uses to extend certain SQL functionalities to non-SQL parts of actions.
*/
package evaluater

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// Evaluater can evaluate arbitrary SQL strings.
// It uses an in-memory, read-only sqlite database to evaluate the expressions.
type Evaluater struct {
	sqlite *sqlite.MemoryConnection
}

// Evaluate evaluates the given expression with the given values.
func (e *Evaluater) Evaluate(expr string, values map[string]any) (any, error) {
	result, err := e.sqlite.Query(expr, values)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	if len(result) != 1 {
		return nil, fmt.Errorf("error evaluating in-line expression: expected 1 row, got %d", len(result))
	}

	if len(result[0]) != 1 {
		return nil, fmt.Errorf("error evaluating in-line expression: expected 1 column, got %d", len(result[0]))
	}

	val, ok := result[0][resultAlias]
	if !ok {
		return nil, fmt.Errorf("error evaluating in-line expression: result not found")
	}

	return val, nil
}

// Close closes the Evaluater.
func (e *Evaluater) Close() error {
	return e.sqlite.Close()
}

// NewEvaluater creates a new Evaluater.
func NewEvaluater() (*Evaluater, error) {
	sql, err := sqlite.OpenReadOnlyMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
	}
	return &Evaluater{
		sqlite: sql,
	}, nil
}

const resultAlias = "result"

// PrepareExpression generates a sqlite query from the given expression, to be used for evaluation.
// It will generate statements as: "SELECT <expr> AS result"
func PrepareExpression(expr tree.Expression) (string, error) {
	switch e := expr.(type) {
	case *tree.ExpressionLiteral, *tree.ExpressionBindParameter, *tree.ExpressionUnary, *tree.ExpressionBinaryComparison, *tree.ExpressionFunction, *tree.ExpressionArithmetic:
		// do nothing
	default:
		return "", fmt.Errorf("unsupported expression type: %T", e)
	}

	stmt := &tree.Select{
		SelectStmt: &tree.SelectStmt{
			SelectCores: []*tree.SelectCore{
				{
					SelectType: tree.SelectTypeAll,
					Columns: []tree.ResultColumn{
						&tree.ResultColumnExpression{
							Expression: expr,
							Alias:      resultAlias,
						},
					},
				},
			},
		},
	}

	return stmt.ToSQL()
}
