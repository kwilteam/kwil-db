package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestCTE_ToSQL(t *testing.T) {
	type fields struct {
		Table   string
		Columns []string
		Select  *tree.SelectCore
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "valid cte",
			fields: fields{
				Table:   "foo",
				Columns: []string{"bar", "baz"},
				Select: &tree.SelectCore{
					SelectCores: []*tree.SimpleSelect{
						{
							SelectType: tree.SelectTypeAll,
							From: &tree.RelationTable{
								Name: "foo",
							},
						},
					},
				},
			},
			want: `"foo" ("bar", "baz") AS (SELECT * FROM "foo")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("CTE.ToSQL() should not have panicked")
					}
				}()
			}

			c := &tree.CTE{
				Table:   tt.fields.Table,
				Columns: tt.fields.Columns,
				Select:  tt.fields.Select,
			}

			got := c.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("CTE.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

var mockCTE = &tree.CTE{
	Table:   "foo",
	Columns: []string{"bar", "baz"},
	Select: &tree.SelectCore{
		SelectCores: []*tree.SimpleSelect{
			{
				From: &tree.RelationTable{
					Name: "foo",
				},
			},
		},
	},
}
