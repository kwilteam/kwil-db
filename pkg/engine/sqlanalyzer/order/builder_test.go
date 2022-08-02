package order_test

import (
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

// a basic builder to remove some testing repetition

type selectBuilder struct {
	cols       []string
	colAliases []string
	from       struct {
		table string
		alias string
	}
	joins []struct {
		table string
		left  string
		right string
		alias string
	}
	union *tree.CompoundOperatorType
}

func Select() SelectBuilder {
	return &selectBuilder{
		cols: []string{},
		from: struct {
			table string
			alias string
		}{},
		joins: []struct {
			table string
			left  string
			right string
			alias string
		}{},
	}
}

type SelectBuilder interface {
	Columns(columns ...string) SelectBuilder
	ColumnAs(as ...string) SelectBuilder
	From(table string, alias ...string) SelectBuilder
	Join(table string, onLeft string, onRight string, alias ...string) SelectBuilder
	Compound(operator tree.CompoundOperatorType) SelectBuilder
	Build() *tree.SelectCore
}

func (s *selectBuilder) Columns(columns ...string) SelectBuilder {
	s.cols = columns
	return s
}

func (s *selectBuilder) ColumnAs(as ...string) SelectBuilder {
	s.colAliases = as
	return s
}

func (s *selectBuilder) From(table string, alias ...string) SelectBuilder {
	s.from.table = table
	if len(alias) > 0 {
		s.from.alias = alias[0]
	}
	return s
}

func (s *selectBuilder) Join(table string, onLeft string, onRight string, alias ...string) SelectBuilder {
	al := ""
	if len(alias) > 0 {
		al = alias[0]
	}

	s.joins = append(s.joins, struct {
		table string
		left  string
		right string
		alias string
	}{table: table, left: onLeft, right: onRight, alias: al})
	return s
}

func (s *selectBuilder) Compound(operator tree.CompoundOperatorType) SelectBuilder {
	s.union = &operator
	return s
}

func (s *selectBuilder) Build() *tree.SelectCore {
	cols := []tree.ResultColumn{}
	if len(s.colAliases) > 0 {
		if len(s.colAliases) != len(s.cols) {
			panic("test selectBuilder: column aliases must be the same length as columns")
		}
	}
	if len(s.cols) == 0 {
		cols = []tree.ResultColumn{&tree.ResultColumnStar{}}
	} else {
		for i, col := range s.cols {
			splitCol := strings.Split(col, ".")

			var tblName string
			var colName string
			if len(splitCol) == 1 {
				colName = splitCol[0]
			} else {
				tblName = splitCol[0]
				colName = splitCol[1]
			}

			if tblName == "*" {
				cols = append(cols, &tree.ResultColumnStar{})
				continue
			}
			if colName == "*" {
				cols = append(cols, &tree.ResultColumnTable{
					TableName: tblName,
				})
				continue
			}

			var colAlias string
			if len(s.colAliases) > 0 {
				colAlias = s.colAliases[i]
			}

			cols = append(cols, &tree.ResultColumnExpression{
				Expression: &tree.ExpressionColumn{
					Table:  tblName,
					Column: colName,
				},
				Alias: colAlias,
			})
		}
	}

	joins := []*tree.JoinPredicate{}

	for _, j := range s.joins {
		leftSide := strings.Split(j.left, ".")
		rightSide := strings.Split(j.right, ".")

		joins = append(joins, &tree.JoinPredicate{
			JoinOperator: &tree.JoinOperator{
				JoinType: tree.JoinTypeJoin,
			},
			Table: &tree.TableOrSubqueryTable{
				Name:  j.table,
				Alias: j.alias,
			},
			Constraint: &tree.ExpressionBinaryComparison{
				Left: &tree.ExpressionColumn{
					Table:  leftSide[0],
					Column: leftSide[1],
				},
				Operator: tree.ComparisonOperatorEqual,
				Right: &tree.ExpressionColumn{
					Table:  rightSide[0],
					Column: rightSide[1],
				},
			},
		})
	}

	sel := &tree.SelectCore{
		Columns: cols,
		From: &tree.FromClause{
			JoinClause: &tree.JoinClause{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  s.from.table,
					Alias: s.from.alias,
				},
				Joins: joins,
			},
		},
	}

	if s.union != nil {
		sel.Compound = &tree.CompoundOperator{
			Operator: *s.union,
		}
	}

	return sel
}
