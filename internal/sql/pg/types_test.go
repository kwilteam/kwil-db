package pg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

	arr2 := []int{1, 2, 3}
	res, err = serializeArray(arr2, 1, func(i int) ([]byte, error) {
		return []byte{byte(i)}, nil
	})
	require.NoError(t, err)

	res3, err := deserializeArray[int64](res, 1, func(b []byte) (any, error) {
		return int(b[0]), nil
	})
	require.NoError(t, err)

	require.EqualValues(t, arr2, res3)
}
