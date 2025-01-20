package types

import (
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodedValue_EdgeCases(t *testing.T) {
	t.Run("encode max int64", func(t *testing.T) {
		ev, err := EncodeValue(int64(9223372036854775807))
		require.NoError(t, err)
		decoded, err := ev.Decode()
		require.NoError(t, err)
		assert.Equal(t, int64(9223372036854775807), decoded)
	})

	t.Run("encode min int64", func(t *testing.T) {
		ev, err := EncodeValue(int64(-9223372036854775808))
		require.NoError(t, err)
		decoded, err := ev.Decode()
		require.NoError(t, err)
		assert.Equal(t, int64(-9223372036854775808), decoded)
	})

	t.Run("encode empty string", func(t *testing.T) {
		ev, err := EncodeValue("")
		require.NoError(t, err)
		decoded, err := ev.Decode()
		require.NoError(t, err)
		assert.Equal(t, "", decoded)
	})

	t.Run("encode empty byte slice", func(t *testing.T) {
		ev, err := EncodeValue([]byte{})
		require.NoError(t, err)
		decoded, err := ev.Decode()
		require.NoError(t, err)
		assert.Equal(t, []byte{}, decoded)
	})

	// THIS IS INCORRECT WITH scientific notation e.g 1e-28
	t.Run("encode decimal with max precision", func(t *testing.T) {
		d, err := decimal.NewFromBigInt(new(big.Int).SetInt64(1), -6)
		require.NoError(t, err)
		ev, err := EncodeValue(d)
		require.NoError(t, err)
		decoded, err := ev.Decode()
		require.NoError(t, err)
		assert.Equal(t, d.String(), decoded.(*decimal.Decimal).String())
	})

	t.Run("encode mixed array types should fail", func(t *testing.T) {
		_, err := EncodeValue([]interface{}{"string", 42})
		assert.Error(t, err)
	})

	t.Run("decode invalid type", func(t *testing.T) {
		ev := &EncodedValue{
			Type: DataType{Name: "invalid_type"},
			Data: [][]byte{{1, 2, 3}},
		}
		_, err := ev.Decode()
		assert.Error(t, err)
	})

	t.Run("decode invalid int length", func(t *testing.T) {
		ev := &EncodedValue{
			Type: DataType{Name: IntType.Name},
			Data: [][]byte{{1, 2, 3}},
		}
		_, err := ev.Decode()
		assert.Error(t, err)
	})

	t.Run("decode invalid uuid length", func(t *testing.T) {
		ev := &EncodedValue{
			Type: DataType{Name: UUIDType.Name},
			Data: [][]byte{{1, 2, 3}},
		}
		_, err := ev.Decode()
		assert.Error(t, err)
	})

	t.Run("encode/decode decimal array", func(t *testing.T) {
		d1, _ := decimal.NewFromString("100")
		d2, _ := decimal.NewFromString("200")
		arr := decimal.DecimalArray{d1, d2}

		ev, err := EncodeValue(arr)
		require.NoError(t, err)

		decoded, err := ev.Decode()
		require.NoError(t, err)

		decodedArr, ok := decoded.(decimal.DecimalArray)
		require.True(t, ok)
		assert.Equal(t, arr[0].String(), decodedArr[0].String())
		assert.Equal(t, arr[1].String(), decodedArr[1].String())
	})
}

func TestDecodeArrayTypes(t *testing.T) {
	t.Run("decode text array", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: TextType.Name, IsArray: true},
			Data: [][]byte{[]byte("hello"), []byte("world")},
		}
		decoded, err := e.Decode()
		require.NoError(t, err)
		assert.Equal(t, []string{"hello", "world"}, decoded)
	})

	t.Run("decode int array", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: IntType.Name, IsArray: true},
			Data: [][]byte{
				binary.BigEndian.AppendUint64(nil, 123),
				binary.BigEndian.AppendUint64(nil, 456),
			},
		}
		decoded, err := e.Decode()
		require.NoError(t, err)
		assert.Equal(t, []int64{123, 456}, decoded)
	})

	t.Run("decode blob array", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: BlobType.Name, IsArray: true},
			Data: [][]byte{{1, 2, 3}, {4, 5, 6}},
		}
		decoded, err := e.Decode()
		require.NoError(t, err)
		assert.Equal(t, [][]byte{{1, 2, 3}, {4, 5, 6}}, decoded)
	})

	t.Run("decode bool array", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: BoolType.Name, IsArray: true},
			Data: [][]byte{{1}, {0}},
		}
		decoded, err := e.Decode()
		require.NoError(t, err)
		assert.Equal(t, []bool{true, false}, decoded)
	})

	t.Run("decode null array error", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: NullType.Name, IsArray: true},
			Data: [][]byte{nil, nil},
		}
		_, err := e.Decode()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot decode array of type 'null'")
	})

	t.Run("decode array with invalid element", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: IntType.Name, IsArray: true},
			Data: [][]byte{[]byte("123"), []byte("invalid")},
		}
		_, err := e.Decode()
		require.Error(t, err)
	})

	t.Run("decode empty array", func(t *testing.T) {
		e := &EncodedValue{
			Type: DataType{Name: TextType.Name, IsArray: true},
			Data: [][]byte{},
		}
		decoded, err := e.Decode()
		require.NoError(t, err)
		assert.Equal(t, []string{}, decoded)
	})
}
