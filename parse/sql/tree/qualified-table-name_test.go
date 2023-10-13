package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestQualifiedTableName_ToSQL(t *testing.T) {
	type fields struct {
		TableName  string
		TableAlias string
		IndexedBy  string
		NotIndexed bool
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
			name: "indexed by",
			fields: fields{
				TableName: "foo",
				IndexedBy: "bar",
			},
			want: `"foo" INDEXED BY "bar"`,
		},
		{
			name: "not indexed",
			fields: fields{
				TableName:  "foo",
				NotIndexed: true,
			},
			want: `"foo" NOT INDEXED`,
		},
		{
			name: "indexed by and not indexed",
			fields: fields{
				TableName:  "foo",
				IndexedBy:  "bar",
				NotIndexed: true,
			},
			wantPanic: true,
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
				IndexedBy:  tt.fields.IndexedBy,
				NotIndexed: tt.fields.NotIndexed,
			}

			got := q.ToSQL()

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("QualifiedTableName.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
