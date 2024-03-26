package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestRelation_ToSQL(t *testing.T) {
	type fields struct {
		Relation tree.Relation
		Schema   string // optional, only for RelationTable
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
				Relation: &tree.RelationTable{
					Name:  "foo",
					Alias: "f",
				},
			},
			want: `"foo" AS "f"`,
		},
		{
			name: "no table name",
			fields: fields{
				Relation: &tree.RelationTable{
					Alias: "f",
				},
			},
			wantPanic: true,
		},
		{
			name: "subquery",
			fields: fields{
				Relation: &tree.RelationSubquery{
					Select: &tree.SelectCore{
						SimpleSelects: []*tree.SimpleSelect{
							{
								SelectType: tree.SelectTypeAll,
								From: &tree.RelationTable{
									Name: "foo",
								},
							},
						},
					},
				},
			},
			want: `(SELECT * FROM "foo")`,
		},
		{
			name: "join",
			fields: fields{
				Relation: &tree.RelationJoin{
					Relation: &tree.RelationTable{
						Name: "foo",
					},
					Joins: []*tree.JoinPredicate{
						{
							JoinOperator: &tree.JoinOperator{
								JoinType: tree.JoinTypeLeft,
								Outer:    true,
							},
							Table: &tree.RelationTable{
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
			want: `"foo" LEFT OUTER JOIN "bar" ON "foo" = "bar"`,
		},
		{
			name: "join clause with one join predicate",
			fields: fields{
				Relation: &tree.RelationJoin{
					Relation: &tree.RelationTable{
						Name:  "table1",
						Alias: "t1",
					},
					Joins: []*tree.JoinPredicate{
						{
							JoinOperator: &tree.JoinOperator{
								JoinType: tree.JoinTypeInner,
							},
							Table: &tree.RelationTable{
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
			},
			want: `"table1" AS "t1" INNER JOIN "table2" AS "t2" ON "t1"."col1" = "t2"."col2"`,
		},
		{
			name: "join clause with greater than one join predicate",
			fields: fields{
				Relation: &tree.RelationJoin{
					Relation: &tree.RelationTable{
						Name:  "table1",
						Alias: "t1",
					},
					Joins: []*tree.JoinPredicate{
						{
							JoinOperator: &tree.JoinOperator{
								JoinType: tree.JoinTypeInner,
							},
							Table: &tree.RelationTable{
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
			},
			wantPanic: true,
		},
		{
			name: "join clause with one side of the join operator not containing a column",
			fields: fields{
				Relation: &tree.RelationJoin{
					Relation: &tree.RelationTable{
						Name:  "table1",
						Alias: "t1",
					},
					Joins: []*tree.JoinPredicate{
						{
							JoinOperator: &tree.JoinOperator{
								JoinType: tree.JoinTypeInner,
							},
							Table: &tree.RelationTable{
								Name:  "table2",
								Alias: "t2",
							},
							Constraint: &tree.ExpressionBinaryComparison{
								Left: &tree.ExpressionColumn{
									Table:  "t1",
									Column: "col1",
								},
								Operator: tree.ComparisonOperatorEqual,
								Right: &tree.ExpressionTextLiteral{
									Value: "value",
								},
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
				Relation: &tree.RelationJoin{
					Relation: &tree.RelationTable{
						Name:  "table1",
						Alias: "t1",
					},
					Joins: []*tree.JoinPredicate{
						{
							JoinOperator: &tree.JoinOperator{
								JoinType: tree.JoinTypeInner,
							},
							Table: &tree.RelationTable{
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
			},
			wantPanic: true,
		},
		{
			name: "schema namespace",
			fields: fields{
				Relation: &tree.RelationTable{
					Name:  "foo",
					Alias: "f",
				},
				Schema: "baz",
			},
			want: `"baz"."foo" AS "f"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Relation.ToSQL() should have panicked")
					}
				}()
			}

			tr := tt.fields.Relation

			if tt.fields.Schema != "" {
				trt := tr.(*tree.RelationTable)
				trt.SetSchema(tt.fields.Schema)
			}

			got := tr.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("Relation.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
