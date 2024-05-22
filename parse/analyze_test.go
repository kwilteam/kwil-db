package parse

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

func Test_Scope(t *testing.T) {
	tbl := &types.Table{}
	s := sqlContext{
		joinedTables: map[string]*types.Table{
			"table1": tbl,
		},
	}

	s.scope()
	s.scope()

	require.EqualValues(t, s.joinedTables, map[string]*types.Table{})

	s.popScope()
	s.popScope()

	require.EqualValues(t, s.joinedTables, map[string]*types.Table{
		"table1": tbl,
	})
}
