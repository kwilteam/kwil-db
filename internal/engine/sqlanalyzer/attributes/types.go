package attributes

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/utils"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// predictReturnType will attempt to predict the return type of an expression.
// If it is ambiguous but is a valid return expression, it will return types.TEXT.
// If it is invalid, it will return an error.
func predictReturnType(expr tree.Expression, tables []*types.Table) (types.DataType, error) {
	w := &returnTypeWalker{
		AstWalker: tree.NewBaseWalker(),
		tables:    tables,
	}

	err := expr.Walk(w)
	if err != nil {
		return types.TEXT, fmt.Errorf("error predicting return type: %w", err)
	}

	if !w.detected {
		return types.TEXT, fmt.Errorf("could not detect return type for expression: %s", expr)
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
	tree.AstWalker
	detected     bool
	detectedType types.DataType
	tables       []*types.Table
}

var _ tree.AstWalker = &returnTypeWalker{}

func (r *returnTypeWalker) EnterExpressionArithmetic(p0 *tree.ExpressionArithmetic) error {
	r.set(types.INT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionBetween(p0 *tree.ExpressionBetween) error {
	r.set(types.INT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionBinaryComparison(p0 *tree.ExpressionBinaryComparison) error {
	r.set(types.INT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionBindParameter(p0 *tree.ExpressionBindParameter) error {
	r.set(types.TEXT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionCase(p0 *tree.ExpressionCase) error {
	r.set(types.TEXT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionCollate(p0 *tree.ExpressionCollate) error {
	r.set(types.TEXT)
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

func (r *returnTypeWalker) EnterExpressionDistinct(p0 *tree.ExpressionDistinct) error {
	r.set(types.INT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionFunction(p0 *tree.ExpressionFunction) error {
	if r.detected {
		return nil
	}

	switch p0.Function {

	// scalars
	case &tree.FunctionABS:
		r.set(types.INT)
	case &tree.FunctionCOALESCE:
		// ambiguous
		r.set(types.TEXT)
	case &tree.FunctionERROR:
		return fmt.Errorf("%w: using function %s", ErrInvalidReturnExpression, p0.Function.Name())
	case &tree.FunctionFORMAT:
		r.set(types.TEXT)
	case &tree.FunctionGLOB:
		r.set(types.INT)
	case &tree.FunctionHEX:
		r.set(types.TEXT)
	case &tree.FunctionIFNULL:
		// ambiguous
		r.set(types.TEXT)
	case &tree.FunctionIIF:
		// ambiguous
		r.set(types.TEXT)
	case &tree.FunctionINSTR:
		r.set(types.INT)
	case &tree.FunctionLENGTH:
		r.set(types.INT)
	case &tree.FunctionLIKE:
		r.set(types.INT)
	case &tree.FunctionLOWER:
		r.set(types.TEXT)
	case &tree.FunctionLTRIM:
		r.set(types.TEXT)
	case &tree.FunctionNULLIF:
		// ambiguous
		r.set(types.TEXT)
	case &tree.FunctionQUOTE:
		r.set(types.TEXT)
	case &tree.FunctionREPLACE:
		r.set(types.TEXT)
	case &tree.FunctionRTRIM:
		r.set(types.TEXT)
	case &tree.FunctionSIGN:
		r.set(types.INT)
	case &tree.FunctionSUBSTR:
		r.set(types.TEXT)
	case &tree.FunctionTRIM:
		r.set(types.TEXT)
	case &tree.FunctionTYPEOF:
		r.set(types.TEXT)
	case &tree.FunctionUNHEX:
		r.set(types.TEXT)
	case &tree.FunctionUNICODE:
		r.set(types.INT)
	case &tree.FunctionUPPER:
		r.set(types.TEXT)

		// aggregates
	case &tree.FunctionCOUNT:
		r.set(types.INT)
	case &tree.FunctionGROUPCONCAT:
		r.set(types.TEXT)
	case &tree.FunctionMAX:
		r.set(types.INT)
	case &tree.FunctionMIN:
		r.set(types.INT)

		// datetime (all return text)
	case &tree.FunctionDATE, &tree.FunctionTIME, &tree.FunctionDATETIME, &tree.FunctionUNIXEPOCH, &tree.FunctionSTRFTIME:
		r.set(types.TEXT)
	default:
		return fmt.Errorf("unknown function: %s", p0.Function)
	}

	return nil
}
func (r *returnTypeWalker) EnterExpressionIsNull(p0 *tree.ExpressionIsNull) error {
	r.set(types.INT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionList(p0 *tree.ExpressionList) error {
	return errReturnExpr(p0)
}

// EnterExpressionLiteral will attempt to detect the type of the literal
func (r *returnTypeWalker) EnterExpressionLiteral(p0 *tree.ExpressionLiteral) error {
	if r.detected {
		return nil
	}

	dataTypes, err := utils.IsLiteral(p0.Value)
	if err != nil {
		return err
	}
	switch dataTypes {
	case types.TEXT:
		r.set(types.TEXT)
	case types.INT:
		r.set(types.INT)
	default:
		return fmt.Errorf("unknown literal type for analyzed relation attribute: %s", dataTypes)
	}

	return nil
}
func (r *returnTypeWalker) EnterExpressionSelect(p0 *tree.ExpressionSelect) error {
	return errReturnExpr(p0)
}
func (r *returnTypeWalker) EnterExpressionStringCompare(p0 *tree.ExpressionStringCompare) error {
	r.set(types.INT)
	return nil
}
func (r *returnTypeWalker) EnterExpressionUnary(p0 *tree.ExpressionUnary) error {
	r.set(types.INT)
	return nil
}

// set sets the detected type if it has not already been set
// since we only want the first detected type
func (r *returnTypeWalker) set(t types.DataType) {
	if r.detected {
		return
	}

	r.detected = true
	r.detectedType = t
}
