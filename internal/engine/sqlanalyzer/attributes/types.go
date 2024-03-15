package attributes

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

// predictReturnType will attempt to predict the return type of an expression.
// If it is ambiguous but is a valid return expression, it will return types.TextType.
// If it is invalid, it will return an error.
func predictReturnType(expr tree.Expression, tables []*types.Table) (*types.DataType, error) {
	w := &returnTypeWalker{
		AstListener: tree.NewBaseListener(),
		tables:      tables,
	}

	err := expr.Walk(w)
	if err != nil {
		return types.TextType, fmt.Errorf("error predicting return type: %w", err)
	}

	if !w.detected {
		return types.IntType, fmt.Errorf("could not detect return type for expression: %s", expr)
	}

	return w.detectedType, nil
}

// ErrInvalidReturnExpression is returned when an expression cannot be used as a result column
var ErrInvalidReturnExpression = fmt.Errorf("expression cannot be used as a result column")

// errReturnExpr is used to return an error when an expression cannot be used as a result column
func errReturnExpr(expr tree.Expression) error {
	return fmt.Errorf("%w: using expression %s", ErrInvalidReturnExpression, expr)
}

type returnTypeWalker struct {
	tree.AstListener
	detected     bool
	detectedType *types.DataType
	tables       []*types.Table
}

var _ tree.AstListener = &returnTypeWalker{}

func (r *returnTypeWalker) EnterExpressionArithmetic(p0 *tree.ExpressionArithmetic) error {
	r.set(types.IntType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionBetween(p0 *tree.ExpressionBetween) error {
	r.set(types.IntType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionBinaryComparison(p0 *tree.ExpressionBinaryComparison) error {
	r.set(types.IntType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionBindParameter(p0 *tree.ExpressionBindParameter) error {
	r.set(types.TextType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionCase(p0 *tree.ExpressionCase) error {
	r.set(types.TextType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionCollate(p0 *tree.ExpressionCollate) error {
	r.set(types.TextType)
	return nil
}

// we need to identify the column type
// there are three potential cases here
// 1. the expression declares the table
//   - we need to search the table for the column to get the data type
//
// 2. the expression does not declare the table, but usedTables is not empty
//   - we need to search the first usedTables table for the column to get the data type, and add the table name to the column
//   - if we can't find the column, we return an error
//
// 3. the expression does not declare the table, and usedTables is empty
//   - we return an error
func (r *returnTypeWalker) EnterExpressionColumn(p0 *tree.ExpressionColumn) error {
	if r.detected {
		return nil
	}

	// case 1
	if p0.Table != "" {
		table, err := findTable(r.tables, p0.Table)
		if err != nil {
			return err
		}

		col, err := findColumn(table.Columns, p0.Column)
		if err != nil {
			return err
		}

		r.set(col.Type)
		return nil
	}

	// case 2
	if len(r.tables) > 0 && r.tables[0] != nil {
		col, err := findColumn(r.tables[0].Columns, p0.Column)
		if err != nil {
			return err
		}

		r.set(col.Type)
		return nil
	}

	// case 3
	return fmt.Errorf(`%w: could not identify column "%s"`, ErrInvalidReturnExpression, p0.Column)

}

// Boolean somewhere?

func (r *returnTypeWalker) EnterExpressionFunction(p0 *tree.ExpressionFunction) error {
	if r.detected {
		return nil
	}

	switch p0.Function {

	// scalars
	case &tree.FunctionABS:
		r.set(types.IntType)
	case &tree.FunctionERROR:
		return fmt.Errorf("%w: using function %s", ErrInvalidReturnExpression, p0.Function.Name())
	case &tree.FunctionFORMAT:
		r.set(types.TextType)
	case &tree.FunctionLENGTH:
		r.set(types.IntType)
	case &tree.FunctionLOWER:
		r.set(types.TextType)
	case &tree.FunctionUPPER:
		r.set(types.TextType)

		// aggregates
	case &tree.FunctionCOUNT:
		r.set(types.IntType)
	case &tree.FunctionSUM:
		r.set(types.IntType)
	default:
		return fmt.Errorf("unknown function: %s", p0.Function)
	}

	return nil
}
func (r *returnTypeWalker) EnterExpressionIs(p0 *tree.ExpressionIs) error {
	r.set(types.IntType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionList(p0 *tree.ExpressionList) error {
	return errReturnExpr(p0)
}

func (r *returnTypeWalker) EnterExpressionTextLiteral(p0 *tree.ExpressionTextLiteral) error {
	r.set(types.TextType)
	return nil
}

func (r *returnTypeWalker) EnterExpressionNumericLiteral(p0 *tree.ExpressionNumericLiteral) error {
	r.set(types.IntType)
	return nil
}

func (r *returnTypeWalker) EnterExpressionBooleanLiteral(p0 *tree.ExpressionBooleanLiteral) error {
	r.set(types.BoolType)
	return nil
}

func (r *returnTypeWalker) EnterExpressionNullLiteral(p0 *tree.ExpressionNullLiteral) error {
	r.set(types.TextType)
	return nil
}

func (r *returnTypeWalker) EnterExpressionBlobLiteral(p0 *tree.ExpressionBlobLiteral) error {
	r.set(types.BlobType)
	return nil
}

func (r *returnTypeWalker) EnterExpressionSelect(p0 *tree.ExpressionSelect) error {
	return errReturnExpr(p0)
}
func (r *returnTypeWalker) EnterExpressionStringCompare(p0 *tree.ExpressionStringCompare) error {
	r.set(types.IntType)
	return nil
}
func (r *returnTypeWalker) EnterExpressionUnary(p0 *tree.ExpressionUnary) error {
	r.set(types.IntType)
	return nil
}

// set sets the detected type if it has not already been set
// since we only want the first detected type
func (r *returnTypeWalker) set(t *types.DataType) {
	if r.detected {
		return
	}

	r.detected = true
	r.detectedType = t
}
