package types_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

func Test_Uint256Math(t *testing.T) {
	// simply testing that the base number is not modified
	a, err := types.Uint256FromString("500")
	require.NoError(t, err)

	b, err := types.Uint256FromString("10000000000")
	require.NoError(t, err)

	c := a.Add(b)
	require.Equal(t, "10000000500", c.String())
	require.Equal(t, "500", a.String())
	require.Equal(t, "10000000000", b.String())

	// go underflow
	_, err = a.Sub(b)
	require.Error(t, err)

	// div without mod
	d, err := types.Uint256FromString("498")
	require.NoError(t, err)

	e := a.Div(d)
	require.Equal(t, "1", e.String())

	// div mod
	f, g := a.DivMod(d)
	require.Equal(t, "1", f.String())
	require.Equal(t, "2", g.String())
}
