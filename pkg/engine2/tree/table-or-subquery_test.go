package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine2/tree"
)

func TestTableOrSubqueryTable_ToSQL(t *testing.T) {
	type fields struct {
		TableOrSubquery tree.TableOrSubquery
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid table",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Name:  "foo",
					Alias: "f",
				},
			},
			want: `"foo" AS "f"`,
		},
		{
			name: "no table name",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryTable{
					Alias: "f",
				},
			},
			wantPanic: true,
		},
		{
			name: "subquery",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubquerySelect{
					Select: &tree.SelectStmt{
						SelectCore: &tree.SelectCore{
							SelectType: tree.SelectTypeAll,
							From: &tree.FromClause{
								JoinClause: &tree.JoinClause{
									TableOrSubquery: &tree.TableOrSubqueryTable{
										Name: "foo",
									},
								},
							},
						},
					},
				},
			},
			want: `(SELECT * FROM "foo")`,
		},
		{
			name: "list",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryList{
					TableOrSubqueries: []tree.TableOrSubquery{
						&tree.TableOrSubqueryTable{
							Name: "foo",
						},
						&tree.TableOrSubquerySelect{
							Select: &tree.SelectStmt{
								SelectCore: &tree.SelectCore{
									SelectType: tree.SelectTypeAll,
									From: &tree.FromClause{
										JoinClause: &tree.JoinClause{
											TableOrSubquery: &tree.TableOrSubqueryTable{
												Name: "foo",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: `("foo", (SELECT * FROM "foo"))`,
		},
		{
			name: "empty list",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryList{},
			},
			wantPanic: true,
		},
		{
			name: "join",
			fields: fields{
				TableOrSubquery: &tree.TableOrSubqueryJoin{
					JoinClause: &tree.JoinClause{
						TableOrSubquery: &tree.TableOrSubqueryTable{
							Name: "foo",
						},
						Joins: []*tree.JoinPredicate{
							{
								JoinOperator: &tree.JoinOperator{
									JoinType: tree.JoinTypeLeft,
									Outer:    true,
								},
								Table: &tree.TableOrSubqueryTable{
									Name: "bar",
								},
								Constraint: &tree.ExpressionBinaryComparison{
									Left:     &tree.ExpressionColumn{Column: "foo"},
									Operator: tree.ComparisonOperatorEqual,
									Right:    &tree.ExpressionColumn{Column: "bar"},
								},
							},
						},
					},
				},
			},
			want: `("foo" LEFT OUTER JOIN "bar" ON "foo" = "bar")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("TableOrSubquery.ToSQL() should have panicked")
					}
				}()
			}

			tr := tt.fields.TableOrSubquery

			got := tr.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("TableOrSubquery.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
