package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func TestAggregateFunc_String(t *testing.T) {
	type fields struct {
		*tree.AggregateFunc
	}
	type args struct {
		exprs []tree.Expression
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		wantPanic bool
	}{
		{
			name: "count fails with no arguments",
			fields: fields{
				AggregateFunc: tree.FunctionCOUNTGetter(nil).(*tree.AggregateFunc),
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
				},
			},
			want: `count(DISTINCT "foo")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("AggregateFunc.String() should have panicked")
					}
				}()
			}

			s := tt.fields.AggregateFunc
			s.SetDistinct(true)
			got := s.ToString(tt.args.exprs...)

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("AggregateFunc.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
