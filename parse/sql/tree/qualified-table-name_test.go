package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestQualifiedTableName_ToSQL(t *testing.T) {
	type fields struct {
		TableName  string
		TableAlias string
		Schema     string // optional
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid table name",
			fields: fields{
				TableName:  "foo",
				TableAlias: "f",
			},
			want: `"foo" AS "f"`,
		},
		{
			name: "no table name",
			fields: fields{
				TableAlias: "f",
			},
			wantPanic: true,
		},
		{
			name: "schema",
			fields: fields{
				TableName: "foo",
				Schema:    "baz",
			},
			want: `"baz"."foo"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("QualifiedTableName.ToSQL() should have panicked")
					}
				}()
			}

			q := &tree.QualifiedTableName{
				TableName:  tt.fields.TableName,
				TableAlias: tt.fields.TableAlias,
			}

			if tt.fields.Schema != "" {
				q.SetSchema(tt.fields.Schema)
			}

			got := q.ToSQL()

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("QualifiedTableName.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
