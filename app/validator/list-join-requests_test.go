package validator

import (
	"encoding/json"
	"testing"

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
						Candidate: types.HexBytes{0x12, 0x34},
						Power:     100,
						ExpiresAt: 900,
						Board:     []types.HexBytes{{0xAB, 0xCD}},
						Approved:  []bool{true},
					},
					{
						Candidate: types.HexBytes{0x56, 0x78},
						Power:     200,
						ExpiresAt: 1000,
						Board:     []types.HexBytes{{0xEF, 0x12}},
						Approved:  []bool{false},
					},
				},
			},
			want: `[{"candidate":"1234","power":100,"expires_at":900,"board":["abcd"],"approved":[true]},{"candidate":"5678","expires_at":1000,"power":200,"board":["ef12"],"approved":[false]}]`,
		},
		{
			name: "empty joins",
			response: respJoinList{
				Joins: []*types.JoinRequest{},
			},
			want: `[]`,
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
						Candidate: types.HexBytes{0x12, 0x34},
						Power:     100,
						ExpiresAt: 1000,
						Board:     []types.HexBytes{{0xAB}},
						Approved:  []bool{true},
					},
				},
			},
			want: "Pending join requests (1 approval needed):\n Candidate                                                        | Power | Approvals | Expiration\n------------------------------------------------------------------+-------+-----------+------------\n 1234 |   100 |         1 | 1000",
		},
		{
			name: "multiple approvals needed",
			response: respJoinList{
				Joins: []*types.JoinRequest{
					{
						Candidate: types.HexBytes{0x12, 0x34},
						Power:     100,
						ExpiresAt: 1000,
						Board:     []types.HexBytes{{0xAB}, {0xCD}, {0xEF}},
						Approved:  []bool{true, false, true},
					},
					{
						Candidate: types.HexBytes{0x56, 0x78},
						Power:     200,
						ExpiresAt: 2000,
						Board:     []types.HexBytes{{0xAB}, {0xCD}, {0xEF}},
						Approved:  []bool{false, false, false},
					},
				},
			},
			want: "Pending join requests (2 approvals needed):\n Candidate                                                        | Power | Approvals | Expiration\n------------------------------------------------------------------+-------+-----------+------------\n 1234 |   100 |         2 | 1000\n 5678 |   200 |         0 | 2000",
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
