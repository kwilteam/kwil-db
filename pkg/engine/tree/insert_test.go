package tree_test

import (
	"kwil/pkg/engine/tree"
	"testing"
)

func TestInsertStatement_ToSql(t *testing.T) {
	type fields struct {
		InsertType      tree.InsertType
		Table           string
		TableAlias      string
		Columns         []string
		Values          [][]tree.InsertExpression
		Upsert          *tree.Upsert
		ReturningClause *tree.ReturningClause
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
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
				Values: [][]tree.InsertExpression{
					{
						&tree.ExpressionLiteral{Value: "barVal"},
						&tree.ExpressionBindParameter{Parameter: "$a"},
					},
					{
						&tree.ExpressionLiteral{Value: "bazVal"},
						&tree.ExpressionBindParameter{Parameter: "$b"},
					},
				},
			},
			want:    `INSERT INTO  "foo"  AS "f"  ("barCol", "bazCol") VALUES ('barVal', $a), ('bazVal', $b);`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &tree.Insert{
				InsertType:      tt.fields.InsertType,
				Table:           tt.fields.Table,
				TableAlias:      tt.fields.TableAlias,
				Columns:         tt.fields.Columns,
				Values:          tt.fields.Values,
				Upsert:          tt.fields.Upsert,
				ReturningClause: tt.fields.ReturningClause,
			}
			got, err := i.ToSql()
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertStatement.ToSql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("InsertStatement.ToSql() = %v, want %v", got, tt.want)
			}
		})
	}
}
