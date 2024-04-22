package utils

import (
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

// GetUsedTables returns the tables that are used or joined in a Join Clause.
// It will search across the base table as well as all joined predicates.
// It will properly scope tables used in subqueries, and not include them in the result.
func GetUsedTables(join tree.Relation) ([]*tree.RelationTable, error) {
	tables := make([]*tree.RelationTable, 0)
	depth := 0 // depth tracks if we are in a subquery or not

	err := join.Walk(&tree.ImplementedListener{
		FuncEnterExpressionSelect: func(p0 *tree.ExpressionSelect) error {
			depth++

			return nil
		},
		FuncExitExpressionSelect: func(p0 *tree.ExpressionSelect) error {
			depth--
			return nil
		},
		FuncEnterRelationTable: func(p0 *tree.RelationTable) error {
			if depth != 0 {
				return nil
			}

			tables = append(tables, &tree.RelationTable{
				Name:  p0.Name,
				Alias: p0.Alias,
			})
			return nil
		},
		FuncEnterRelationSubquery: func(p0 *tree.RelationSubquery) error {
			if depth != 0 {
				return nil
			}
			depth++ // we add depth since we do not want to index extra information from the subquery

			// simply call the name and alias the alias of the subquery
			tables = append(tables, &tree.RelationTable{
				Name:  p0.Alias,
				Alias: p0.Alias,
			})
			return nil
		},
		FuncExitRelationSubquery: func(p0 *tree.RelationSubquery) error {
			depth--
			return nil
		},
	})

	return tables, err
}
