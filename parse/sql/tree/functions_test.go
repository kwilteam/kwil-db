package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func Test_sqlFunction_String(t *testing.T) {
	type fields struct {
		Function tree.AnySQLFunction
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
			name: "valid function",
			fields: fields{
				Function: tree.AnySQLFunction{
					FunctionName: "abs",
					Min:          1,
					Max:          1,
				},
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
				},
			},
			want: `abs("foo")`,
		},
		{
			name: "valid function with multiple args",
			fields: fields{
				Function: tree.AnySQLFunction{
					FunctionName: "coalesce",
					Min:          2,
				},
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
					&tree.ExpressionColumn{Column: "bar"},
				},
			},
			want: `coalesce("foo", "bar")`,
		},
		{
			name: "valid function with too few args",
			fields: fields{
				Function: tree.AnySQLFunction{
					FunctionName: "coalesce",
					Min:          2,
				},
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
				},
			},
			wantPanic: true,
		},
		{
			name: "valid function with too many args",
			fields: fields{
				Function: tree.AnySQLFunction{
					FunctionName: "coalesce",
					Max:          2,
				},
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionColumn{Column: "foo"},
					&tree.ExpressionColumn{Column: "bar"},
					&tree.ExpressionColumn{Column: "baz"},
				},
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &tree.AnySQLFunction{
				FunctionName: tt.fields.Function.FunctionName,
				Min:          tt.fields.Function.Min,
				Max:          tt.fields.Function.Max,
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("sqlFunction.String() should have panicked")
					}
				}()
			}
			got := s.ToString(tt.args.exprs...)
			if tt.wantPanic {
				return
			}
			b := compareIgnoringWhitespace(got, tt.want)
			if b == false {
				t.Errorf("sqlFunction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ScalarFunction_String(t *testing.T) {
	type fields struct {
		Function *tree.ScalarFunction
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
			name: "format function",
			fields: fields{
				Function: tree.FunctionFORMATGetter(nil).(*tree.ScalarFunction),
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{Value: "'Hello, %s'"},
					&tree.ExpressionLiteral{Value: "'World'"},
				},
			},
			want: "format('Hello, %s', 'World')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &tree.AnySQLFunction{
				FunctionName: tt.fields.Function.FunctionName,
				Min:          tt.fields.Function.Min,
				Max:          tt.fields.Function.Max,
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("sqlFunction.String() should have panicked")
					}
				}()
			}
			got := s.ToString(tt.args.exprs...)
			if tt.wantPanic {
				return
			}
			b := compareIgnoringWhitespace(got, tt.want)
			if b == false {
				t.Errorf("sqlFunction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
