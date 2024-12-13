package types_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestPayload struct {
	val string
}

func (tp *TestPayload) MarshalBinary() ([]byte, error) {
	return []byte(tp.val), nil
}

func (tp *TestPayload) UnmarshalBinary(data []byte) error {
	tp.val = string(data)
	return nil
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
		{"kv pair payload", types.PayloadTypeExecute, true},
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

func TestValidatorVoteBodyMarshalUnmarshal(t *testing.T) {
	voteBody := &types.ValidatorVoteBodies{
		Events: []*types.VotableEvent{
			{
				Type: "emptydata",
				Body: []byte(""),
			},
			{
				Type: "test",
				Body: []byte("test"),
			},
			{
				Type: "test2",
				Body: []byte("random large data, random large data,random large data,random large data,random large data,random large data,random large data,random large data,random large data,random large data,random large data,random large data,random large data,"),
			},
		},
	}

	data, err := voteBody.MarshalBinary()
	require.NoError(t, err)

	voteBody2 := &types.ValidatorVoteBodies{}
	err = voteBody2.UnmarshalBinary(data)
	require.NoError(t, err)

	require.NotNil(t, voteBody2)
	require.NotNil(t, voteBody2.Events)
	require.Len(t, voteBody2.Events, 3)

	require.Equal(t, voteBody.Events, voteBody2.Events)
}
