package validator

import (
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

func Test_respValJoinStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		response respValJoinStatus
		want     string
	}{
		{
			name: "basic marshal",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: []byte{0x12, 0x34},
					Power:     100,
					Board:     []types.HexBytes{{0xAB, 0xCD}, {0xEF, 0x12}},
					Approved:  []bool{true, false},
				},
			},
			want: `{"candidate":"1234","power":100,"board":["abcd","ef12"],"approved":[true,false]}`,
		},
		{
			name: "empty board",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: []byte{0xFF},
					Power:     50,
					Board:     []types.HexBytes{},
					Approved:  []bool{},
				},
			},
			want: `{"candidate":"ff","power":50,"board":[],"approved":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(&tt.response)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(got))
		})
	}
}

func Test_respValJoinStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		response respValJoinStatus
		want     string
	}{
		{
			name: "all approvals",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: []byte{0x12, 0x34},
					Power:     1000,
					ExpiresAt: 5000,
					Board:     []types.HexBytes{{0xAB, 0xCD}, {0xEF, 0x12}, {0x56, 0x78}},
					Approved:  []bool{true, true, true},
				},
			},
			want: "Candidate: 1234\nRequested Power: 1000\nExpiration Height: 5000\n3 Approvals Received (2 needed):\nValidator abcd, approved\nValidator ef12, approved\nValidator 5678, approved\n",
		},
		{
			name: "mixed approvals",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: []byte{0xFF},
					Power:     500,
					ExpiresAt: 1000,
					Board:     []types.HexBytes{{0x11, 0x22}, {0x33, 0x44}},
					Approved:  []bool{true, false},
				},
			},
			want: "Candidate: ff\nRequested Power: 500\nExpiration Height: 1000\n1 Approvals Received (2 needed):\nValidator 1122, approved\nValidator 3344, not approved\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.response.MarshalText()
			require.NoError(t, err)
			require.Equal(t, tt.want, string(got))
		})
	}
}
