package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

func TestInsert_ToSQL(t *testing.T) {
	type fields struct {
		CTE        []*tree.CTE
		InsertStmt *tree.InsertCore
		Schema     string
	}
	tests := []struct {
		name    string
		fields  fields
		wantStr string
		wantErr bool
	}{
		{
			name: "valid insert",
			fields: fields{
				CTE: []*tree.CTE{
					mockCTE,
				},
				InsertStmt: &tree.InsertCore{
					InsertType: tree.InsertTypeInsert,
					Table:      "foo",
					Columns:    []string{"bar", "baz"},
					Values: [][]tree.Expression{
						{
							&tree.ExpressionTextLiteral{Value: "barVal"},
							&tree.ExpressionBindParameter{Parameter: "$a"},
						},
						{
							&tree.ExpressionTextLiteral{Value: "bazVal"},
							&tree.ExpressionBindParameter{Parameter: "$b"},
						},
					},
					Upsert: &tree.Upsert{
						ConflictTarget: &tree.ConflictTarget{
							IndexedColumns: []string{"bar", "baz"},
						},
						Type: tree.UpsertTypeDoNothing,
					},
					ReturningClause: &tree.ReturningClause{
						Returned: []*tree.ReturningClauseColumn{
							{
								All: true,
							},
						},
					},
				},
			},
			wantStr: `WITH ` + mockCTE.ToSQL() + ` INSERT INTO "foo" ("bar", "baz") VALUES ('barVal', $a), ('bazVal', $b) ON CONFLICT ("bar", "baz") DO NOTHING RETURNING *;`,
		},
		{
			name: "insert to namespaced",
			fields: fields{
				InsertStmt: &tree.InsertCore{
					InsertType: tree.InsertTypeInsert,
					Table:      "bar",
					Columns:    []string{"baz"},
					Values: [][]tree.Expression{
						{
							&tree.ExpressionTextLiteral{Value: "bazVal"},
						},
					},
				},
				Schema: "public",
			},
			wantStr: `INSERT INTO "public"."bar" ("baz") VALUES ('bazVal');`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins := &tree.InsertStmt{
				CTE:  tt.fields.CTE,
				Core: tt.fields.InsertStmt,
			}

			if tt.fields.Schema != "" {
				ins.Core.SetSchema(tt.fields.Schema)
			}

			gotStr, err := tree.SafeToSQL(ins)
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertStmt.ToSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareIgnoringWhitespace(gotStr, tt.wantStr) {
				t.Errorf("InsertStmt.ToSQL() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestInsertStatement_ToSql(t *testing.T) {
	type fields struct {
		InsertType      tree.InsertType
		Table           string
		TableAlias      string
		Columns         []string
		Values          [][]tree.Expression
		Upsert          *tree.Upsert
		ReturningClause *tree.ReturningClause
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid insert",
			fields: fields{
				InsertType: tree.InsertTypeInsert,
				Table:      "foo",
				TableAlias: "f",
				Columns: []string{
					"barCol",
					"bazCol",
				},
				Values: [][]tree.Expression{
					{
						&tree.ExpressionTextLiteral{Value: "barVal"},
						&tree.ExpressionBindParameter{Parameter: "$a"},
					},
					{
						&tree.ExpressionTextLiteral{Value: "bazVal"},
						&tree.ExpressionBindParameter{Parameter: "$b"},
					},
				},
				Upsert: &tree.Upsert{
					ConflictTarget: &tree.ConflictTarget{
						IndexedColumns: []string{"barCol", "bazCol"},
						Where: &tree.ExpressionBinaryComparison{
							Left: &tree.ExpressionTextLiteral{
								Value: "barVal",
							},
							Operator: tree.ComparisonOperatorEqual,
							Right: &tree.ExpressionBindParameter{
								Parameter: "$a",
							},
						},
					},
					Type: tree.UpsertTypeDoUpdate,
					Updates: []*tree.UpdateSetClause{
						{
							Columns: []string{"barCol"},
							Expression: &tree.ExpressionBindParameter{
								Parameter: "$a",
							},
						},
					},
					Where: &tree.ExpressionBinaryComparison{
						Left: &tree.ExpressionTextLiteral{
							Value: "barVal",
						},
						Operator: tree.ComparisonOperatorEqual,
						Right: &tree.ExpressionBindParameter{
							Parameter: "$a",
						},
					},
				},
				ReturningClause: &tree.ReturningClause{
					Returned: []*tree.ReturningClauseColumn{
						{
							All: true,
						},
					},
				},
			},
			want:      `INSERT INTO  "foo"  AS "f"  ("barCol", "bazCol") VALUES ('barVal', $a), ('bazVal', $b) ON CONFLICT ("barCol", "bazCol") WHERE 'barVal' = $a DO UPDATE SET "barCol" = $a WHERE 'barVal' = $a RETURNING *`,
			wantPanic: false,
		},
		{
			name: "insert without table",
			fields: fields{
				InsertType: tree.InsertTypeInsert,
				Columns: []string{
					"barCol",
					"bazCol",
				},
				Values: [][]tree.Expression{
					{
						&tree.ExpressionTextLiteral{Value: "barVal"},
						&tree.ExpressionBindParameter{Parameter: "$a"},
					},
					{
						&tree.ExpressionTextLiteral{Value: "bazVal"},
						&tree.ExpressionBindParameter{Parameter: "$b"},
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "insert without values",
			fields: fields{
				InsertType: tree.InsertTypeInsert,
				Table:      "foo",
				Columns: []string{
					"barCol",
					"bazCol",
				},
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("InsertStatement.ToSql() should have panicked")
					}
				}()
			}

			i := &tree.InsertCore{
				InsertType:      tt.fields.InsertType,
				Table:           tt.fields.Table,
				TableAlias:      tt.fields.TableAlias,
				Columns:         tt.fields.Columns,
				Values:          tt.fields.Values,
				Upsert:          tt.fields.Upsert,
				ReturningClause: tt.fields.ReturningClause,
			}
			got := i.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("InsertStatement.ToSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
