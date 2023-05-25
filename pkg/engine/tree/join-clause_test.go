package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/tree"
)

func TestJoinClause_ToSQL(t *testing.T) {
	type fields struct {
		TableOrSubquery tree.TableOrSubquery
		Joins           []*tree.JoinPredicate
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "join clause with one join predicate",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  "table1",
					Alias: "t1",
				},
				Joins: []*tree.JoinPredicate{
					{
						JoinOperator: &tree.JoinOperator{
							JoinType: tree.JoinTypeInner,
						},
						Table: &tree.TableOrSubqueryTable{
							Name:  "table2",
							Alias: "t2",
						},
						Constraint: &tree.ExpressionBinaryComparison{
							Left: &tree.ExpressionColumn{
								Table:  "t1",
								Column: "col1",
							},
							Operator: tree.ComparisonOperatorEqual,
							Right: &tree.ExpressionColumn{
								Table:  "t2",
								Column: "col2",
							},
						},
					},
				},
			},
			want: `"table1" AS "t1" INNER JOIN "table2" AS "t2" ON "t1"."col1" = "t2"."col2"`,
		},
		{
			name: "join clause with greater than one join predicate",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  "table1",
					Alias: "t1",
				},
				Joins: []*tree.JoinPredicate{
					{
						JoinOperator: &tree.JoinOperator{
							JoinType: tree.JoinTypeInner,
						},
						Table: &tree.TableOrSubqueryTable{
							Name:  "table2",
							Alias: "t2",
						},
						Constraint: &tree.ExpressionBinaryComparison{
							Left: &tree.ExpressionColumn{
								Table:  "t1",
								Column: "col1",
							},
							Operator: tree.ComparisonOperatorGreaterThan,
							Right: &tree.ExpressionColumn{
								Table:  "t2",
								Column: "col2",
							},
						},
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "join clause with one side of the join operator not containing a column",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  "table1",
					Alias: "t1",
				},
				Joins: []*tree.JoinPredicate{
					{
						JoinOperator: &tree.JoinOperator{
							JoinType: tree.JoinTypeInner,
						},
						Table: &tree.TableOrSubqueryTable{
							Name:  "table2",
							Alias: "t2",
						},
						Constraint: &tree.ExpressionBinaryComparison{
							Left: &tree.ExpressionColumn{
								Table:  "t1",
								Column: "col1",
							},
							Operator: tree.ComparisonOperatorEqual,
							Right: &tree.ExpressionLiteral{
								Value: "'value'",
							},
						},
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "join clause with only 1 column as the condition",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  "table1",
					Alias: "t1",
				},
				Joins: []*tree.JoinPredicate{
					{
						JoinOperator: &tree.JoinOperator{
							JoinType: tree.JoinTypeInner,
						},
						Table: &tree.TableOrSubqueryTable{
							Name:  "table2",
							Alias: "t2",
						},
						Constraint: &tree.ExpressionColumn{
							Table:  "t1",
							Column: "col1",
						},
					},
				},
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &tree.JoinClause{
				TableOrSubquery: tt.fields.TableOrSubquery,
				Joins:           tt.fields.Joins,
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("JoinClause.ToSQL() should have panicked")
					}
				}()
			}

			got := j.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("JoinClause.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
