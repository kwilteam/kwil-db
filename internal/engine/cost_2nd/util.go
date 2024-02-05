package cost_2nd

import (
	"strings"
	"unicode"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// removeWhitespace removes all whitespace characters from a string.
func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // skip this rune
		}
		return r
	}, s)
}

func exprToField(expr tree.Expression, input LogicalPlan) *field {
	switch t := expr.(type) {
	case *tree.ExpressionColumn:
		for _, field := range input.Schema().fields {
			if field.ColName == t.Column {
				return field
			}
		}
	case *tree.ExpressionLiteral:
		return &field{
			ColName: t.Value,
			//Type:    t.Type,
			Type: "text",
		}
	case *tree.ExpressionFunction:
		// TODO: determine the return type of function
		var retType string
		switch t.Function.Name() {
		case "abs", "count", "min", "max":
			retType = "int"
		default:
			retType = "text"
		}
		return &field{
			ColName:         t.Function.Name(),
			OriginalColName: t.Function.Name(),
			Type:            retType,
		}
	case *tree.ExpressionCase:
		return &field{
			ColName: "case",
		}
	case *tree.ExpressionArithmetic:
		return &field{
			ColName: t.Operator.String(),
			Type:    "int",
		}
	case *tree.ExpressionBetween:
		return &field{
			ColName: "between",
			Type:    "int", //bool
		}
	case *tree.ExpressionBinaryComparison:
		return &field{
			ColName: t.Operator.String(),
			Type:    "int", //bool
		}
	case *tree.ExpressionIsNull:
		return &field{
			ColName: "isnull",
			Type:    "int", //bool
		}
	case *tree.ExpressionDistinct:
		return &field{
			ColName: "distinct",
			Type:    "int", //bool
		}
	case *tree.ExpressionStringCompare:
		return &field{
			ColName: t.Operator.String(),
			Type:    "int", //bool
		}
	case *tree.ExpressionCollate:
		// NOTE: probably will be removed
		return &field{
			ColName: "collate",
			Type:    "text",
		}
	case *tree.ExpressionUnary:
		return &field{
			ColName: t.Operator.String(),
			Type:    "int", //bool
		}
	case *tree.ExpressionBindParameter:
		//prefix := t.Text()
		//if t.Parameter[0] == '@' {
		//
		//}
		return &field{
			ColName: "bind",
			Type:    "text", // NOTE: unknown
		}
	case *tree.ExpressionSelect:
		return &field{
			ColName: "select",
			Type:    "text", // NOTE: unknown
		}
	case *tree.ExpressionList:
		// TODO:
		fs := make([]*field, len(t.Expressions))
		for i, expr := range t.Expressions {
			fs[i] = exprToField(expr, input)
		}
	default:
		panic("not implemented")
	}

	return nil
}
