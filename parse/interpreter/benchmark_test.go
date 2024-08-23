package interpreter

import (
	"context"
	"math"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
)

func BenchmarkLoops(b *testing.B) {
	tests := []struct {
		name string
		proc *types.Procedure
		args []any
	}{
		// {
		// 	name: "simple_loop",
		// 	proc: &types.Procedure{
		// 		Name: "simple_loop",
		// 		Body: `
		// 		$result int[];
		// 		for $i in 1..100 {
		// 			$result := array_append($result, $i*2);
		// 		}
		// 		return $result;
		// 		`,
		// 	},
		// },
		{
			name: "test_loop",
			args: []any{1, 1000000},
			proc: &types.Procedure{
				Name: "loop",
				Body: `
				$res := 0;
				for $i in $start..$end {
					$res := $res + $i;
				}
				return $res;
				`,
				Parameters: []*types.ProcedureParameter{
					{Name: "$start", Type: types.IntType},
					{Name: "$end", Type: types.IntType},
				},
				Returns: &types.ProcedureReturn{
					IsTable: false,
					Fields: []*types.NamedType{
						{Name: "res", Type: types.IntType},
					},
				},
			},
		},
	}

	ctx := context.Background()
	schema := &types.Schema{
		Name:       "test_schema",
		Procedures: []*types.Procedure{},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			schema.Procedures = []*types.Procedure{tt.proc}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Run(ctx, tt.proc, schema, tt.args, math.MaxInt64, ZeroCostTable())
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
