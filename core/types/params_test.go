package types

import (
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/stretchr/testify/require"
)

func TestSetParamNames(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantPanic bool
	}{
		{
			name:      "valid struct with all fields",
			input:     NetworkParameters{},
			wantPanic: false,
		},
		{
			name: "missing json tag",
			input: struct {
				Leader string
			}{},
			wantPanic: true,
		},
		{
			name: "unknown field",
			input: struct {
				UnknownField string `json:"unknown"`
			}{},
			wantPanic: true,
		},
		{
			name: "unset params (partial fields)",
			input: struct {
				Leader string `json:"leader"`
			}{},
			wantPanic: true,
		},
		{
			name:      "nil input",
			input:     nil,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("setParamNames() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			setParamNames(tt.input)

			if !tt.wantPanic {
				// Verify the parameter names were set correctly
				if ParamNameLeader != "leader" {
					t.Errorf("ParamNameLeader = %v, want %v", ParamNameLeader, "leader")
				}
				if ParamNameMaxBlockSize != "max_block_size" {
					t.Errorf("ParamNameMaxBlockSize = %v, want %v", ParamNameMaxBlockSize, "max_block_size")
				}
				if ParamNameJoinExpiry != "join_expiry" {
					t.Errorf("ParamNameJoinExpiry = %v, want %v", ParamNameJoinExpiry, "join_expiry")
				}
				if ParamNameDisabledGasCosts != "disabled_gas_costs" {
					t.Errorf("ParamNameDisabledGasCosts = %v, want %v", ParamNameDisabledGasCosts, "disabled_gas_costs")
				}
				if ParamNameMaxVotesPerTx != "max_votes_per_tx" {
					t.Errorf("ParamNameMaxVotesPerTx = %v, want %v", ParamNameMaxVotesPerTx, "max_votes_per_tx")
				}
				if ParamNameMigrationStatus != "migration_status" {
					t.Errorf("ParamNameMigrationStatus = %v, want %v", ParamNameMigrationStatus, "migration_status")
				}
			}
		})
	}
}

func TestParamUpdatesMarshalBinary(t *testing.T) {
	_, pub, err := crypto.GenerateSecp256k1Key(nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		updates ParamUpdates
		wantErr bool
	}{
		{
			name:    "empty updates",
			updates: ParamUpdates{},
			wantErr: false,
		},
		{
			name: "all parameter types",
			updates: ParamUpdates{
				ParamNameLeader:           PublicKey{pub},
				ParamNameMaxBlockSize:     int64(1000),
				ParamNameJoinExpiry:       Duration(10 * time.Second),
				ParamNameDisabledGasCosts: true,
				ParamNameMaxVotesPerTx:    int64(10),
				ParamNameMigrationStatus:  MigrationStatus("pending"),
			},
			wantErr: false,
		},
		{
			name: "invalid leader type",
			updates: ParamUpdates{
				ParamNameLeader: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid numeric type",
			updates: ParamUpdates{
				ParamNameMaxBlockSize: "1000",
			},
			wantErr: true,
		},
		{
			name: "invalid boolean type",
			updates: ParamUpdates{
				ParamNameDisabledGasCosts: "true",
			},
			wantErr: true,
		},
		{
			name: "invalid migration status type",
			updates: ParamUpdates{
				ParamNameMigrationStatus: 123,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.updates.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			var decoded ParamUpdates
			err = decoded.UnmarshalBinary(data)
			if err != nil {
				t.Errorf("UnmarshalBinary() error = %v", err)
				return
			}

			if len(decoded) != len(tt.updates) {
				t.Errorf("Decoded length mismatch: got %v, want %v", len(decoded), len(tt.updates))
			}

			for k, v := range tt.updates {
				decodedVal, exists := decoded[k]
				if !exists {
					t.Errorf("Missing key in decoded: %v", k)
					continue
				}
				if !reflect.DeepEqual(v, decodedVal) {
					t.Errorf("Value mismatch for key %v: got %v, want %v", k, decodedVal, v)
				}
			}
		})
	}
}

func TestParamUpdatesUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "invalid number of updates",
			input:   []byte{255, 255, 255, 255},
			wantErr: true,
		},
		{
			name:    "truncated input",
			input:   []byte{1, 0, 0, 0, 6, 0, 0, 0, 'l', 'e', 'a', 'd', 'e', 'r'},
			wantErr: true,
		},
		{
			name:    "invalid parameter name",
			input:   []byte{1, 0, 0, 0, 7, 0, 0, 0, 'i', 'n', 'v', 'a', 'l', 'i', 'd'},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updates ParamUpdates
			err := updates.UnmarshalBinary(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeUpdates(t *testing.T) {
	pub0, err := crypto.UnmarshalSecp256k1PublicKey([]byte{0x2, 0xe0, 0x9d, 0x79, 0x32, 0xde, 0xf1, 0x1d, 0x82, 0x72, 0xdd, 0x3b, 0x58, 0x9d, 0xf8, 0xb1, 0xcf, 0x7a, 0xff, 0xb0, 0x41, 0x50, 0x19, 0x4f, 0xc2, 0x28, 0xf8, 0x17, 0xae, 0xba, 0xb2, 0xc9, 0xda})
	if err != nil {
		t.Fatal(err)
	}
	// acct0 := acctIDForPubKey(pub0)
	pub1, err := crypto.UnmarshalSecp256k1PublicKey([]byte{0x3, 0x16, 0xb4, 0x4c, 0xab, 0xfb, 0xc, 0xc, 0xa1, 0x3b, 0x58, 0xc4, 0x69, 0x3f, 0x71, 0xd8, 0xd0, 0xf1, 0x6e, 0xcb, 0x16, 0xe9, 0xb6, 0xed, 0xd3, 0xa2, 0x23, 0x74, 0xef, 0x38, 0xc7, 0xf0, 0xb})
	if err != nil {
		t.Fatal(err)
	}
	// acct1 := acctIDForPubKey(pub1)
	tests := []struct {
		name    string
		np      *NetworkParameters
		updates ParamUpdates
		wantErr bool
		verify  func(*testing.T, *NetworkParameters)
	}{
		{
			name: "update single field",
			np: &NetworkParameters{
				Leader: PublicKey{pub0},
			},
			updates: ParamUpdates{
				ParamNameLeader: PublicKey{pub1},
			},
			wantErr: false,
			verify: func(t *testing.T, np *NetworkParameters) {
				if !pub1.Equals(np.Leader.PublicKey) {
					t.Errorf("Leader not updated correctly, got %v want %v", np.Leader, pub1)
				}
			},
		},
		{
			name: "update multiple fields",
			np: &NetworkParameters{
				MaxBlockSize:     1000,
				DisabledGasCosts: false,
			},
			updates: ParamUpdates{
				ParamNameMaxBlockSize:     int64(2000),
				ParamNameDisabledGasCosts: true,
			},
			wantErr: false,
			verify: func(t *testing.T, np *NetworkParameters) {
				if np.MaxBlockSize != 2000 {
					t.Errorf("MaxBlockSize not updated correctly, got %v want %v", np.MaxBlockSize, 2000)
				}
				if !np.DisabledGasCosts {
					t.Errorf("DisabledGasCosts not updated correctly, got %v want true", np.DisabledGasCosts)
				}
			},
		},
		{
			name: "wrong type assertion",
			np:   &NetworkParameters{},
			updates: ParamUpdates{
				ParamNameMaxBlockSize: "not an int64",
			},
			wantErr: true,
		},
		{
			name: "nil network parameters",
			np:   nil,
			updates: ParamUpdates{
				ParamNameLeader: PublicKey{pub0},
			},
			wantErr: true,
		},
		{
			name: "migration status update",
			np: &NetworkParameters{
				MaxBlockSize:    123,
				MigrationStatus: MigrationStatus("pending"),
			},
			updates: ParamUpdates{
				ParamNameMigrationStatus: MigrationStatus("completed"),
			},
			wantErr: false,
			verify: func(t *testing.T, np *NetworkParameters) {
				if np.MigrationStatus != "completed" {
					t.Errorf("MigrationStatus not updated correctly, got %v want completed", np.MigrationStatus)
				}
				if np.MaxBlockSize != 123 {
					t.Errorf("MaxBlockSize modified when updating migration status")
				}
			},
		},
		{
			name: "invalid migration status type",
			np:   &NetworkParameters{},
			updates: ParamUpdates{
				ParamNameMigrationStatus: 123,
			},
			wantErr: true,
		},
		{
			name: "update with empty updates map",
			np: &NetworkParameters{
				Leader: PublicKey{pub0},
			},
			updates: ParamUpdates{},
			wantErr: false,
			verify: func(t *testing.T, np *NetworkParameters) {
				if !pub0.Equals(np.Leader.PublicKey) {
					t.Errorf("Parameters modified when no updates provided")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeUpdates(tt.np, tt.updates)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeUpdates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, tt.np)
			}
		})
	}
}

func TestParamUpdatesMerge(t *testing.T) {
	pub0, err := crypto.UnmarshalSecp256k1PublicKey([]byte{0x2, 0xe0, 0x9d, 0x79, 0x32, 0xde, 0xf1, 0x1d, 0x82, 0x72, 0xdd, 0x3b, 0x58, 0x9d, 0xf8, 0xb1, 0xcf, 0x7a, 0xff, 0xb0, 0x41, 0x50, 0x19, 0x4f, 0xc2, 0x28, 0xf8, 0x17, 0xae, 0xba, 0xb2, 0xc9, 0xda})
	if err != nil {
		t.Fatal(err)
	}
	// acct0 := acctIDForPubKey(pub0)
	pub1, err := crypto.UnmarshalSecp256k1PublicKey([]byte{0x3, 0x16, 0xb4, 0x4c, 0xab, 0xfb, 0xc, 0xc, 0xa1, 0x3b, 0x58, 0xc4, 0x69, 0x3f, 0x71, 0xd8, 0xd0, 0xf1, 0x6e, 0xcb, 0x16, 0xe9, 0xb6, 0xed, 0xd3, 0xa2, 0x23, 0x74, 0xef, 0x38, 0xc7, 0xf0, 0xb})
	if err != nil {
		t.Fatal(err)
	}
	// acct1 := acctIDForPubKey(pub1)

	tests := []struct {
		name     string
		base     ParamUpdates
		other    ParamUpdates
		wantErr  bool // other is invalid or not
		expected ParamUpdates
	}{
		{
			name: "merge into empty base",
			base: ParamUpdates{},
			other: ParamUpdates{
				ParamNameLeader:       PublicKey{pub0},
				ParamNameMaxBlockSize: int64(5000),
			},
			expected: ParamUpdates{
				ParamNameLeader:       PublicKey{pub0},
				ParamNameMaxBlockSize: int64(5000),
			},
		},
		{
			name: "merge empty other",
			base: ParamUpdates{
				ParamNameLeader:       PublicKey{pub0},
				ParamNameMaxBlockSize: int64(5000),
			},
			other: ParamUpdates{},
			expected: ParamUpdates{
				ParamNameLeader:       PublicKey{pub0},
				ParamNameMaxBlockSize: int64(5000),
			},
		},
		{
			name: "override existing values",
			base: ParamUpdates{
				ParamNameLeader:           PublicKey{pub0},
				ParamNameMaxBlockSize:     int64(5000),
				ParamNameDisabledGasCosts: true,
			},
			other: ParamUpdates{
				ParamNameLeader:          PublicKey{pub1},
				ParamNameMaxBlockSize:    int64(6000),
				ParamNameMigrationStatus: MigrationStatus("completed"),
			},
			expected: ParamUpdates{
				ParamNameLeader:           PublicKey{pub1},
				ParamNameMaxBlockSize:     int64(6000),
				ParamNameDisabledGasCosts: true,
				ParamNameMigrationStatus:  MigrationStatus("completed"),
			},
		},
		{
			name: "invalid updates",
			base: ParamUpdates{},
			other: ParamUpdates{
				ParamNameMaxVotesPerTx:    "bad",
				ParamNameDisabledGasCosts: 1.21,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdates(tt.other)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateUpdates() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			tt.base.Merge(tt.other)
			if !reflect.DeepEqual(tt.base, tt.expected) {
				t.Errorf("Merge() result = %v, want %v", tt.base, tt.expected)
			}
		})
	}
}

func TestPublicKeyJSON(t *testing.T) {
	tests := []struct {
		name    string
		pk      PublicKey
		json    string
		wantErr bool
	}{
		{
			name:    "unmarshal invalid hex string",
			json:    `{"type":"public_key","key":"XYZ"}`,
			wantErr: true,
		},
		{
			name:    "unmarshal invalid json structure",
			json:    `{"type":"public_key"}`,
			wantErr: true,
		},
		{
			name:    "unmarshal invalid key type",
			json:    `{"type":"invalid_type","key":"0123"}`,
			wantErr: true,
		},
		{
			name:    "unmarshal malformed json",
			json:    `{"type":,"key":"0123"}`,
			wantErr: true,
		},
		{
			name:    "unmarshal empty json",
			json:    `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				data, err := tt.pk.MarshalJSON()
				require.NoError(t, err)
				require.Equal(t, tt.json, string(data))

				var decoded PublicKey
				err = decoded.UnmarshalJSON([]byte(tt.json))
				require.NoError(t, err)
			} else {
				var decoded PublicKey
				err := decoded.UnmarshalJSON([]byte(tt.json))
				require.Error(t, err)
			}
		})
	}
}

func TestParamUpdatesUnmarshalJSON(t *testing.T) {
	pubBts, err := hex.DecodeString("03642dcd0d9b1821ddf4097c442a300e4aa1593800d3358583ea554271965d792d")
	require.NoError(t, err)

	pub, err := crypto.UnmarshalSecp256k1PublicKey(pubBts)
	require.NoError(t, err)

	tests := []struct {
		name    string
		json    string
		want    ParamUpdates
		wantErr bool
	}{
		{
			name: "all parameter types",
			json: `{
				"leader": {"type":"secp256k1","key":"03642dcd0d9b1821ddf4097c442a300e4aa1593800d3358583ea554271965d792d"},
				"max_block_size": 5000,
				"join_expiry": 3600,
				"disabled_gas_costs": true,
				"max_votes_per_tx": 100,
				"migration_status": "in_progress"
			}`,
			want: ParamUpdates{
				ParamNameLeader:           PublicKey{pub},
				ParamNameMaxBlockSize:     int64(5000),
				ParamNameJoinExpiry:       int64(3600),
				ParamNameDisabledGasCosts: true,
				ParamNameMaxVotesPerTx:    int64(100),
				ParamNameMigrationStatus:  MigrationStatus("in_progress"),
			},
			wantErr: false,
		},
		{
			name: "invalid max_block_size type",
			json: `{
				"max_block_size": "5000"
			}`,
			wantErr: true,
		},
		{
			name: "invalid join_expiry type",
			json: `{
				"join_expiry": true
			}`,
			wantErr: true,
		},
		{
			name: "invalid max_votes_per_tx type",
			json: `{
				"max_votes_per_tx": 3.14
			}`,
			wantErr: true,
		},
		{
			name: "invalid disabled_gas_costs type",
			json: `{
				"disabled_gas_costs": "yes"
			}`,
			wantErr: true,
		},
		{
			name: "invalid migration_status type",
			json: `{
				"migration_status": 123
			}`,
			wantErr: true,
		},
		{
			name:    "malformed json object",
			json:    `{"max_block_size": 1000`,
			wantErr: true,
		},
		{
			name: "empty object",
			json: `{}`,
			want: ParamUpdates{},
		},
		{
			name: "null values",
			json: `{
				"max_block_size": null,
				"disabled_gas_costs": null
			}`,
			wantErr: true,
		},
		{
			name: "mixed valid and invalid fields",
			json: `{
				"max_block_size": 1000,
				"disabled_gas_costs": "invalid",
				"join_expiry": 3600
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pu ParamUpdates
			err := pu.UnmarshalJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(pu, tt.want) {
				t.Errorf("UnmarshalJSON() got = %v, want %v", pu, tt.want)
			}
		})
	}
}
