package tree_test

/*
func TestIndexedColumn_ToSQL(t *testing.T) {
	type fields struct {
		Column     string
		Expression tree.Expression
		Collation  tree.CollationType
		OrderType  tree.OrderType
	}
	tests := []struct {
		name      string
		fields    fields
		want      string
		wantPanic bool
	}{
		{
			name: "column and expression cannot both be empty",
			fields: fields{
				Column:     "",
				Expression: nil,
				Collation:  tree.CollationTypeBinary,
				OrderType:  tree.OrderTypeNone,
			},
			wantPanic: true,
		},
		{
			name: "column and expression cannot both be set",
			fields: fields{
				Column:     "column",
				Expression: &tree.ExpressionLiteral{"expression"},
				Collation:  tree.CollationTypeBinary,
				OrderType:  tree.OrderTypeNone,
			},
			wantPanic: true,
		},
		{
			name: "valid indexed-column with column",
			fields: fields{
				Column:    "col1",
				Collation: tree.CollationTypeRTrim,
				OrderType: tree.OrderTypeAsc,
			},
			want: `"col1" COLLATE RTRIM ASC`,
		},
		{
			name: "valid indexed-column with expression",
			fields: fields{
				Expression: &tree.ExpressionLiteral{"expr"},
				Collation:  tree.CollationTypeRTrim,
				OrderType:  tree.OrderTypeAsc,
			},
			want: "'expr' COLLATE RTRIM ASC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("IndexedColumn.ToSQL() panicked with %v", r)
					}
				}
			}()

			i := &tree.IndexedColumn{
				Column:     tt.fields.Column,
				Expression: tt.fields.Expression,
				Collation:  tt.fields.Collation,
				OrderType:  tt.fields.OrderType,
			}
			got := i.ToSQL()
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("IndexedColumn.ToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
