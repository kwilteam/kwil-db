package tree_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/tree"
)

func TestDateTimeFunction_String(t *testing.T) {
	type fields struct {
		Function tree.DateTimeFunction
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
			name: "date fails with no arguments",
			fields: fields{
				Function: tree.FunctionDATE,
			},
			args:      args{},
			wantPanic: true,
		},
		{
			name: "testing date doesn't work with 'now",
			fields: fields{
				Function: tree.FunctionDATE,
			},
			args: args{
				exprs: []tree.Expression{&tree.ExpressionLiteral{"'now'"}},
			},
			wantPanic: true,
		},
		{
			name: "testing date works with a single argument",
			fields: fields{
				Function: tree.FunctionDATE,
			},
			args: args{
				exprs: []tree.Expression{&tree.ExpressionLiteral{"'06-06-2023'"}},
			},
			want: "date('06-06-2023')",
		},
		{
			name: "testing localtime modifier doesn't work",
			fields: fields{
				Function: tree.FunctionDATE,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'06-06-2023'"},
					&tree.ExpressionLiteral{"'localtime'"},
				},
			},
			wantPanic: true,
		},
		{
			name: "testing modifier whitespace doesn't matter",
			fields: fields{
				Function: tree.FunctionDATE,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'06-06-2023'"},
					&tree.ExpressionLiteral{"'+1 day'"},
				},
			},
			want: "date('06-06-2023', '+1 day')",
		},
		{
			name: "using a floating point number in an otherwise valid date() call",
			fields: fields{
				Function: tree.FunctionDATE,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'06-06-2023'"},
					&tree.ExpressionLiteral{"'+1.3 day'"},
				},
			},
			wantPanic: true,
		},
		{
			name: "strftime using 'now'",
			fields: fields{
				Function: tree.FunctionSTRFTIME,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'%Y-%m-%d %H:%M:%S'"},
					&tree.ExpressionLiteral{"'now'"},
				},
			},
			wantPanic: true,
		},
		{
			name: "strftime with all valid modifiers')",
			fields: fields{
				Function: tree.FunctionSTRFTIME,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'%Y-%m-%d'"},
					&tree.ExpressionLiteral{"'2003-03-03'"},
					&tree.ExpressionLiteral{"'+1 days'"},
					&tree.ExpressionLiteral{"'+10 years'"},
					&tree.ExpressionLiteral{"'-1 months'"},
					&tree.ExpressionLiteral{"'-1 hours'"},
					&tree.ExpressionLiteral{"'-1 minutes'"},
					&tree.ExpressionLiteral{"'start of month'"},
					&tree.ExpressionLiteral{"'start of year'"},
					&tree.ExpressionLiteral{"'start of day'"},
					&tree.ExpressionLiteral{"'weekday 3'"},
					&tree.ExpressionLiteral{"'unixepoch'"},
				},
			},
			want: "strftime('%Y-%m-%d', '2003-03-03', '+1 days', '+10 years', '-1 months', '-1 hours', '-1 minutes', 'start of month', 'start of year', 'start of day', 'weekday 3', 'unixepoch')",
		},
		{
			name: "stfrtime should work with no modifiers",
			fields: fields{
				Function: tree.FunctionSTRFTIME,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'%Y-%m-%d'"},
					&tree.ExpressionLiteral{"'2003-03-03'"},
				},
			},
			want: "strftime('%Y-%m-%d', '2003-03-03')",
		},
		{
			name: "stfrtime should work with modifiers",
			fields: fields{
				Function: tree.FunctionSTRFTIME,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'%Y-%m-%d'"},
					&tree.ExpressionLiteral{"'2003-03-03'"},
					&tree.ExpressionLiteral{"'+1 days'"},
				},
			},
			want: "strftime('%Y-%m-%d', '2003-03-03', '+1 days')",
		},
		{
			name: "stfrtime fails with 1 argument",
			fields: fields{
				Function: tree.FunctionSTRFTIME,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'%Y-%m-%d'"},
				},
			},
			wantPanic: true,
		},
		{
			name: "invalid format substitution",
			fields: fields{
				Function: tree.FunctionSTRFTIME,
			},
			args: args{
				exprs: []tree.Expression{
					&tree.ExpressionLiteral{"'BRUH%Y-%m-%d-%s'"},
					&tree.ExpressionLiteral{"'2003-03-03'"},
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
						t.Errorf("DateTimeFunction.String() should have panicked")
					}
				}()
			}

			got := tt.fields.Function.String(tt.args.exprs...)
			if tt.wantPanic {
				return
			}

			if !compareIgnoringWhitespace(got, tt.want) {
				t.Errorf("DateTimeFunction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
