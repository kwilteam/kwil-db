package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// IsLiteral detects if the passed string is convertable to a literal.
// It returns the type of the literal, or an error if it is not a literal.
func IsLiteral(literal string) (types.DataType, error) {
	if strings.HasPrefix(literal, "'") && strings.HasSuffix(literal, "'") {
		return types.TEXT, nil
	}

	if strings.EqualFold(literal, "true") || strings.EqualFold(literal, "false") {
		return types.BOOL, nil
	}

	if strings.EqualFold(literal, "null") {
		return types.NULL, nil
	}

	_, err := strconv.Atoi(literal)
	if err != nil {
		return types.NULL, fmt.Errorf("invalid literal: could not detect literal type: %s", literal)
	}

	return types.INT, nil
}

// GetUsedTables returns the tables that are used or joined in a Join Clause.
// It will search across the base table as well as all joined predicates.
// It will properly scope tables used in subqueries, and not include them in the result.
func GetUsedTables(join *tree.JoinClause) ([]*tree.TableOrSubqueryTable, error) {
	tables := make([]*tree.TableOrSubqueryTable, 0)
	depth := 0 // depth tracks if we are in a subquery or not

	err := join.Accept(&tree.ImplementedWalker{
		FuncEnterExpressionSelect: func(p0 *tree.ExpressionSelect) error {
			depth++

			return nil
		},
		FuncExitExpressionSelect: func(p0 *tree.ExpressionSelect) error {
			depth--
			return nil
		},
		FuncEnterTableOrSubqueryTable: func(p0 *tree.TableOrSubqueryTable) error {
			if depth != 0 {
				return nil
			}

			tables = append(tables, &tree.TableOrSubqueryTable{
				Name:  p0.Name,
				Alias: p0.Alias,
			})
			return nil
		},
		FuncEnterTableOrSubquerySelect: func(p0 *tree.TableOrSubquerySelect) error {
			if depth != 0 {
				return nil
			}
			depth++ // we add depth since we do not want to index extra information from the subquery

			// simply call the name and alias the alias of the subquery
			tables = append(tables, &tree.TableOrSubqueryTable{
				Name:  p0.Alias,
				Alias: p0.Alias,
			})
			return nil
		},
		FuncExitTableOrSubquerySelect: func(p0 *tree.TableOrSubquerySelect) error {
			depth--
			return nil
		},
	})

	return tables, err
}
