package types_test

import (
	"kwil/node/types/serialize"
	"kwil/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestPayload struct {
	val string
}

func (tp *TestPayload) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(tp.val)
}

func (tp *TestPayload) UnmarshalBinary(data serialize.SerializedData) error {
	return serialize.Decode(data, &tp.val)
}

func (tp *TestPayload) Type() types.PayloadType {
	return "testPayload"
}

func init() {
	types.RegisterPayload("testPayload")
}

func TestValidPayload(t *testing.T) {
	testcases := []struct {
		name  string
		pt    types.PayloadType
		valid bool
	}{
		{"kv pair payload", types.PayloadTypeKV, true},
		{"registered payload", "testPayload", true},
		{"invalid payload", types.PayloadType("unknown"), false},
	}

	for _, tc := range testcases {
		if got := tc.pt.Valid(); got != tc.valid {
			t.Errorf("Expected %v to be %v, got %v", tc.pt, tc.valid, got)
		}
	}
}

func TestMarshalUnmarshalPayload(t *testing.T) {
	tp := &TestPayload{"test"}
	data, err := tp.MarshalBinary()
	require.NoError(t, err)

	var tp2 TestPayload
	err = tp2.UnmarshalBinary(data)
	require.NoError(t, err)

	assert.Equal(t, tp.val, tp2.val)
}
