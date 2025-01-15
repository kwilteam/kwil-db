package validator

import (
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

func Test_respJoinList_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		response respJoinList
		want     string
	}{
		{
			name: "multiple joins",
			response: respJoinList{
				Joins: []*types.JoinRequest{
					{
						Candidate: types.NodeKey{
							PubKey: []byte{0x12, 0x34},
							Type:   crypto.KeyTypeEd25519,
						},
						Power:     100,
						ExpiresAt: 900,
						Board:     []types.NodeKey{{PubKey: []byte{0xAB, 0xCD}, Type: crypto.KeyTypeSecp256k1}},
						Approved:  []bool{true},
					},
					{
						Candidate: types.NodeKey{
							PubKey: []byte{0x56, 0x78},
							Type:   crypto.KeyTypeSecp256k1,
						},
						Power:     200,
						ExpiresAt: 1000,
						Board:     []types.NodeKey{{PubKey: []byte{0xEF, 0x12}, Type: crypto.KeyTypeSecp256k1}},
						Approved:  []bool{false},
					},
				},
			},
			want: `[{"candidate":{"pubkey":"1234","type":1},"power":100,"expires_at":900,"board":[{"pubkey":"abcd","type":0}],"approved":[true]},{"candidate":{"pubkey":"5678","type":0},"power":200,"expires_at":1000,"board":[{"pubkey":"ef12","type":0}],"approved":[false]}]`,
		},
		{
			name: "empty joins",
			response: respJoinList{
				Joins: []*types.JoinRequest{},
			},
			want: `[]`,
		},
		{
			name: "single join",
			response: respJoinList{
				Joins: []*types.JoinRequest{
					{
						Candidate: types.NodeKey{
							PubKey: []byte{0x12, 0x34},
							Type:   crypto.KeyTypeEd25519,
						},
						Power:     150,
						ExpiresAt: 1200,
						Board:     []types.NodeKey{{PubKey: []byte{0xAB, 0xCD}, Type: crypto.KeyTypeSecp256k1}},
						Approved:  []bool{true, false},
					},
				},
			},
			want: `[{"candidate":{"pubkey":"1234","type":1},"power":150,"expires_at":1200,"board":[{"pubkey":"abcd","type":0}],"approved":[true,false]}]`,
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

func Test_respJoinList_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		response respJoinList
		want     string
	}{
		{
			name: "empty list",
			response: respJoinList{
				Joins: []*types.JoinRequest{},
			},
			want: "No pending join requests",
		},
		{
			name: "single approval needed",
			response: respJoinList{
				Joins: []*types.JoinRequest{
					{
						Candidate: types.NodeKey{
							PubKey: []byte{0x12, 0x34},
							Type:   crypto.KeyTypeEd25519,
						},
						Power:     100,
						ExpiresAt: 1000,
						Board:     []types.NodeKey{{PubKey: []byte{0xAB}}},
						Approved:  []bool{true},
					},
				},
			},
			want: "Pending join requests (1 approval needed):\n Candidate                                                        | Power | Approvals | Expiration\n------------------------------------------------------------------+-------+-----------+------------\n NodeKey{pubkey = 1234, keyType = ed25519} |   100 |         1 | 1000",
		},
		{
			name: "multiple approvals needed",
			response: respJoinList{
				Joins: []*types.JoinRequest{
					{
						Candidate: types.NodeKey{
							PubKey: []byte{0x12, 0x34},
							Type:   crypto.KeyTypeEd25519,
						},
						Power:     100,
						ExpiresAt: 1000,
						Board: []types.NodeKey{
							{PubKey: []byte{0xAB}},
							{PubKey: []byte{0xCD}},
							{PubKey: []byte{0xEF}},
						},
						Approved: []bool{true, false, true},
					},
					{
						Candidate: types.NodeKey{
							PubKey: []byte{0x56, 0x78},
						},
						Power:     200,
						ExpiresAt: 2000,
						Board: []types.NodeKey{
							{PubKey: []byte{0xAB}},
							{PubKey: []byte{0xCD}},
							{PubKey: []byte{0xEF}},
						},
						Approved: []bool{false, false, false},
					},
				},
			},
			want: "Pending join requests (2 approvals needed):\n Candidate                                                        | Power | Approvals | Expiration\n------------------------------------------------------------------+-------+-----------+------------\n NodeKey{pubkey = 1234, keyType = ed25519} |   100 |         2 | 1000\n NodeKey{pubkey = 5678, keyType = secp256k1} |   200 |         0 | 2000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.response.MarshalText()
			require.NoError(t, err)
			require.Equal(t, string(got), tt.want)
		})
	}
}
