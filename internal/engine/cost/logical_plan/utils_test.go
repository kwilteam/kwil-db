package logical_plan

import (
	"reflect"
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/stretchr/testify/assert"
)

func TestSplitConjunction(t *testing.T) {
	type args struct {
		expr LogicalExpr
	}

	t1 := datatypes.TableRefUnqualified("t1")

	tests := []struct {
		name string
		args args
		want []LogicalExpr
	}{
		{
			name: "1 level AND",
			args: args{
				expr: And(
					Column(t1, "a"),
					Column(t1, "b"),
				),
			},
			want: []LogicalExpr{
				Column(t1, "a"),
				Column(t1, "b"),
			},
		},
		{
			name: "2 level AND",
			args: args{
				expr: And(
					Column(t1, "a"),
					And(
						Column(t1, "b"),
						Column(t1, "c"),
					),
				),
			},
			want: []LogicalExpr{
				Column(t1, "a"),
				Column(t1, "b"),
				Column(t1, "c"),
			},
		},
		{
			name: "with alias",
			args: args{
				expr: And(
					Alias(Column(t1, "a"), "a"),
					Alias(Column(t1, "b"), "b"),
				),
			},
			want: []LogicalExpr{
				Column(t1, "a"),
				Column(t1, "b"),
			},
		},
		{
			name: "with binary expr",
			args: args{
				expr: And(
					Column(t1, "a"),
					Eq(Column(t1, "b"), LiteralNumeric(1)),
				),
			},
			want: []LogicalExpr{
				Column(t1, "a"),
				Eq(Column(t1, "b"), LiteralNumeric(1)),
			},
		},
		{
			name: "no conjunction",
			args: args{
				expr: Eq(Column(t1, "a"), LiteralNumeric(1)),
			},
			want: []LogicalExpr{
				Eq(Column(t1, "a"), LiteralNumeric(1)),
			},
		},
		{
			name: "no conjunction with alias",
			args: args{
				expr: Alias(Eq(Column(t1, "a"), LiteralNumeric(1)), "a"),
			},
			want: []LogicalExpr{
				Eq(Column(t1, "a"), LiteralNumeric(1)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitConjunction(tt.args.expr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitConjunction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConjunction(t *testing.T) {
	type args struct {
		exprs []LogicalExpr
	}
	t1 := datatypes.TableRefUnqualified("t1")

	tests := []struct {
		name     string
		args     args
		wantExpr LogicalExpr
	}{
		{
			name: "1 level AND",
			args: args{
				exprs: []LogicalExpr{
					Column(t1, "a"),
					Column(t1, "b"),
				},
			},
			wantExpr: And(
				Column(t1, "a"),
				Column(t1, "b"),
			),
		},
		{
			name: "2 level AND",
			args: args{
				exprs: []LogicalExpr{
					Column(t1, "a"),
					Column(t1, "b"),
					Column(t1, "c"),
				},
			},
			wantExpr: And(
				And(Column(t1, "a"),
					Column(t1, "b")),
				Column(t1, "c"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantExpr, Conjunction(tt.args.exprs...))
		})
	}
}
