package pggenerate_test

import (
	"testing"

	pggenerate "github.com/kwilteam/kwil-db/internal/engine/pg_generate"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/require"
)

func Test_PgGenerate(t *testing.T) {
	type testcase struct {
		name    string
		sql     string
		want    string
		params  []string
		wantErr bool
	}

	tests := []testcase{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parse.ParseSQL()

			got, ps, err := pggenerate.GenerateSQL(tt.sql, true, "kwil")
			if err != nil {
				require.Equal(t, tt.wantErr, true)
				return
			} else {
				require.Equal(t, tt.wantErr, false)
			}

			require.Equal(t, tt.want, got)
			require.EqualValues(t, tt.params, ps)
		})
	}
}
