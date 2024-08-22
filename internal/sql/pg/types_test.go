package pg

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// mustDecimal panics if the string cannot be converted to a decimal.
func mustDecimal(s string) *decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustParseUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// mustUint256 panics if the string cannot be converted to a Uint256.
func mustUint256(s string) *types.Uint256 {
	u, err := types.Uint256FromString(s)
	if err != nil {
		panic(err)
	}
	return u
}

func Test_ArrayEncodeDecode(t *testing.T) {
	arr := []string{"a", "b", "c"}
	res, err := serializeArray(arr, 4, func(s string) ([]byte, error) {
		return []byte(s), nil
	})
	require.NoError(t, err)

	res2, err := deserializeArray[string](res, 4, func(b []byte) (any, error) {
		return string(b), nil
	})
	require.NoError(t, err)

	require.EqualValues(t, arr, res2)

	arr2 := []int64{1, 2, 3}
	res, err = serializeArray(arr2, 1, func(i int64) ([]byte, error) {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(i))
		return buf, nil
	})
	require.NoError(t, err)

	res3, err := deserializeArray[int64](res, 1, func(b []byte) (any, error) {
		return int64(binary.LittleEndian.Uint64(b)), nil
	})
	require.NoError(t, err)

	require.EqualValues(t, arr2, res3)
}
