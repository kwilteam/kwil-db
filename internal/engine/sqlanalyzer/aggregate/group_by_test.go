package aggregate_test

import (
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/aggregate"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func Test_GroupBy(t *testing.T) {
	type testCase struct {
		name    string
		query   *tree.SelectCore
		wantErr error
	}

	testCases := []testCase{
		{
			name: "aggregate queries that do encapsulate (otherwise) bare columns in an aggregate function work",
			query: Select().
				Column("a").
				ResultAggregate(&tree.ExpressionFunction{
					Function: &tree.FunctionCOUNT,
					Inputs: []tree.Expression{
						&tree.ExpressionColumn{
							Column: "b",
						},
					},
				}).
				From("table").
				GroupBy("a").
				Build(),
			wantErr: nil,
		},
		{
			name: "aggregate queries that do not encapsulate bare columns in an aggregate function fail",
			query: Select().
				Column("a").
				ResultAggregate(&tree.ExpressionFunction{
					Function: &tree.FunctionCOUNT,
					Inputs: []tree.Expression{
						&tree.ExpressionColumn{
							Column: "b",
						},
					},
				}).
				From("table").
				GroupBy("b").
				Build(),
			wantErr: aggregate.ErrResultSetContainsBareColumn,
		},
		{
			name: "aggregate functions with columns used in positional arguments that are not the first arg fail",
			query: Select().
				ResultAggregate(&tree.ExpressionFunction{
					Function: &tree.FunctionCOUNT,
					Inputs: []tree.Expression{
						&tree.ExpressionColumn{
							Column: "a",
						},
						&tree.ExpressionColumn{
							Column: "b",
						},
					},
				}).
				From("table").
				Build(),
			wantErr: aggregate.ErrAggregateFuncHasInvalidPosArg,
		},
		{
			name: "GROUP BY predicates can only include bare columns (no math, functions, etc.)",
			query: Select().
				Column("a").
				From("table").
				GroupByExpr(
					&tree.ExpressionCollate{
						Expression: &tree.ExpressionColumn{
							Column: "a",
						},
						Collation: "collation",
					},
				).
				Build(),
			wantErr: aggregate.ErrGroupByContainsInvalidExpr,
		},
		{
			name: "HAVING predicates can only use columns in the GROUP BY clause (success case)",
			query: Select().
				Column("a").
				ResultAggregate(
					&tree.ExpressionFunction{
						Function: &tree.FunctionCOUNT,
						Inputs: []tree.Expression{
							&tree.ExpressionColumn{
								Column: "b",
							},
						},
					},
				).
				From("table").
				GroupBy("a").
				Having(&tree.ExpressionCollate{
					Expression: &tree.ExpressionColumn{
						Column: "a",
					},
					Collation: "collation",
				}).
				Build(),
			wantErr: nil,
		},
		{
			name: "HAVING predicates can only use columns in the GROUP BY clause (failure case)",
			query: Select().
				Column("a").
				From("table").
				GroupBy("a").
				Having(&tree.ExpressionCollate{
					Expression: &tree.ExpressionColumn{
						Column: "b",
					},
					Collation: "collation",
				}).
				Build(),
			wantErr: aggregate.ErrHavingClauseContainsUngroupedColumn,
		},
		{
			name: "aggregate functions cannot contain subqueries",
			query: Select().
				ResultAggregate(
					&tree.ExpressionFunction{
						Function: &tree.FunctionCOUNT,
						Inputs: []tree.Expression{
							&tree.ExpressionSelect{
								Select: &tree.SelectStmt{
									SelectCores: []*tree.SelectCore{
										Select().
											Column("a").
											From("table").
											Build(),
									},
								},
							},
						},
					},
				).
				From("table").
				Build(),
			wantErr: aggregate.ErrAggregateFuncContainsSubquery,
		},
		{
			name: "aggregate function does not need a column in p1",
			query: Select().
				ResultAggregate(&tree.ExpressionFunction{
					Function: &tree.FunctionCOUNT,
					Inputs:   []tree.Expression{},
				}).
				From("table").
				Build(),
			wantErr: nil,
		},
		{
			name:    "empty statement does not error",
			query:   &tree.SelectCore{},
			wantErr: nil,
		},
		{
			name: "select * fails",
			query: Select().
				From("table").
				GroupBy("a").
				Build(),
			wantErr: aggregate.ErrAggregateQueryContainsSelectAll,
		},
		{
			name: "select tbl.* fails",
			query: Select().
				ColumnResultRaw(&tree.ResultColumnTable{
					TableName: "table",
				}).
				From("table").
				GroupBy("a").
				Build(),
			wantErr: aggregate.ErrAggregateQueryContainsSelectAll,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			walker := aggregate.NewGroupByWalker()

			err := tc.query.Walk(walker)
			if err != nil && (tc.wantErr == nil) {
				t.Errorf("unexpected error: %v", err)
			}

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected error: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}

type queryBuilder struct {
	columns []tree.ResultColumn

	table struct {
		name  string
		alias string
	}

	groupBys []tree.Expression
	having   tree.Expression
}

func Select() QueryBuilder {
	return &queryBuilder{
		columns: []tree.ResultColumn{},
	}
}

type QueryBuilder interface {
	Column(name string, alias ...string) QueryBuilder
	ColumnResultRaw(expr tree.ResultColumn) QueryBuilder
	ResultAggregate(functionCall *tree.ExpressionFunction, alias ...string) QueryBuilder
	From(table string, alias ...string) QueryBuilder
	GroupBy(columns ...string) QueryBuilder
	GroupByExpr(expr tree.Expression) QueryBuilder
	Having(expr tree.Expression) QueryBuilder
	Build() *tree.SelectCore
}

func (q *queryBuilder) Column(name string, alias ...string) QueryBuilder {
	var aliasName string
	if len(alias) > 0 {
		aliasName = alias[0]
	}

	q.columns = append(q.columns, &tree.ResultColumnExpression{
		Expression: &tree.ExpressionColumn{
			Column: name,
		},
		Alias: aliasName,
	})

	return q
}

func (q *queryBuilder) ColumnResultRaw(expr tree.ResultColumn) QueryBuilder {
	q.columns = append(q.columns, expr)
	return q
}

func (q *queryBuilder) ResultAggregate(functionCall *tree.ExpressionFunction, alias ...string) QueryBuilder {
	var aliasName string
	if len(alias) > 0 {
		aliasName = alias[0]
	}

	q.columns = append(q.columns, &tree.ResultColumnExpression{
		Expression: functionCall,
		Alias:      aliasName,
	})

	return q
}

func (q *queryBuilder) From(name string, alias ...string) QueryBuilder {
	var aliasName string
	if len(alias) > 0 {
		aliasName = alias[0]
	}

	q.table = struct {
		name  string
		alias string
	}{
		name:  name,
		alias: aliasName,
	}

	return q
}

func (q *queryBuilder) GroupBy(columns ...string) QueryBuilder {
	for _, column := range columns {
		q.groupBys = append(q.groupBys, &tree.ExpressionColumn{
			Column: column,
		})
	}

	return q
}

func (q *queryBuilder) GroupByExpr(expr tree.Expression) QueryBuilder {
	q.groupBys = append(q.groupBys, expr)
	return q
}

func (q *queryBuilder) Having(expr tree.Expression) QueryBuilder {
	q.having = expr
	return q
}

func (q *queryBuilder) Build() *tree.SelectCore {
	cols := []tree.ResultColumn{
		&tree.ResultColumnStar{},
	}
	if len(q.columns) > 0 {
		cols = q.columns
	}

	return &tree.SelectCore{
		Columns: cols,
		From: &tree.FromClause{
			JoinClause: &tree.JoinClause{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  q.table.name,
					Alias: q.table.alias,
				},
			},
		},
		GroupBy: &tree.GroupBy{
			Expressions: q.groupBys,
			Having:      q.having,
		},
	}
}
