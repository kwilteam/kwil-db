package validator

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
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
					Candidate: &types.AccountID{
						Identifier: []byte{0x12, 0x34},
						KeyType:    crypto.KeyTypeSecp256k1,
					},
					Power: 100,
					Board: []*types.AccountID{
						{Identifier: []byte{0xEF, 0x12}},
					},
					Approved: []bool{true, false},
				},
			},
			want: `{"candidate":{"identifier":"1234","key_type":"secp256k1"},"power":100,"board":[{"identifier":"ef12","key_type":""}],"approved":[true,false]}`,
		},
		{
			name: "empty board",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: &types.AccountID{
						Identifier: []byte{0xFF},
						KeyType:    crypto.KeyTypeSecp256k1,
					},
					Power:    50,
					Board:    []*types.AccountID{},
					Approved: []bool{},
				},
			},
			want: `{"candidate":{"identifier":"ff","key_type":"secp256k1"},"power":50,"board":[],"approved":[]}`,
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
	now := time.Now()
	nowStr := now.String()
	tests := []struct {
		name     string
		response respValJoinStatus
		want     string
	}{
		{
			name: "all approvals",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: &types.AccountID{Identifier: []byte{0x12, 0x34}, KeyType: crypto.KeyTypeSecp256k1},
					Power:     1000,
					ExpiresAt: now,
					Board: []*types.AccountID{
						{Identifier: []byte{0xAB, 0xCD}, KeyType: crypto.KeyTypeSecp256k1},
						{Identifier: []byte{0xEF, 0x12}, KeyType: crypto.KeyTypeSecp256k1},
						{Identifier: []byte{0x56, 0x78}, KeyType: crypto.KeyTypeSecp256k1},
					},
					Approved: []bool{true, true, true},
				},
			},
			want: "Candidate: AccountID{identifier = 1234, keyType = secp256k1}\nRequested Power: 1000\nExpiration Timestamp: " + nowStr + "\n3 Approvals Received (2 needed):\nValidator AccountID{identifier = abcd, keyType = secp256k1}, approved\nValidator AccountID{identifier = ef12, keyType = secp256k1}, approved\nValidator AccountID{identifier = 5678, keyType = secp256k1}, approved\n",
		},
		{
			name: "mixed approvals",
			response: respValJoinStatus{
				Data: &types.JoinRequest{
					Candidate: &types.AccountID{Identifier: []byte{0xFF}, KeyType: crypto.KeyTypeSecp256k1},
					Power:     500,
					ExpiresAt: now,
					Board: []*types.AccountID{
						{Identifier: []byte{0x11, 0x22}, KeyType: crypto.KeyTypeSecp256k1},
						{Identifier: []byte{0x33, 0x44}, KeyType: crypto.KeyTypeSecp256k1},
					},
					Approved: []bool{true, false},
				},
			},
			want: "Candidate: AccountID{identifier = ff, keyType = secp256k1}\nRequested Power: 500\nExpiration Timestamp: " + nowStr + "\n1 Approvals Received (2 needed):\nValidator AccountID{identifier = 1122, keyType = secp256k1}, approved\nValidator AccountID{identifier = 3344, keyType = secp256k1}, not approved\n",
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
