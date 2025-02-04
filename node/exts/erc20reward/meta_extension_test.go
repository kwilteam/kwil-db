package erc20reward

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSyncedRewardData_MarshalBinary(t *testing.T) {
	tests := []struct {
		name             string
		input            syncedRewardData
		expectedAddr     common.Address
		expectedDecimals int64
	}{
		{
			name: "Valid data positive decimals",
			input: syncedRewardData{
				Erc20Address:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Erc20Decimals: 18,
			},
			expectedAddr:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
			expectedDecimals: 18,
		},
		{
			name: "Valid data negative decimals",
			input: syncedRewardData{
				Erc20Address:  common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				Erc20Decimals: -10,
			},
			expectedAddr:     common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			expectedDecimals: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.input.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary returned an error: %v", err)
			}
			if len(data) != 28 {
				t.Fatalf("expected data length 28, got %d", len(data))
			}
			// Verify that the first 20 bytes match the expected address.
			if !bytes.Equal(data[:20], tt.expectedAddr.Bytes()) {
				t.Errorf("address bytes mismatch: got %x, want %x", data[:20], tt.expectedAddr.Bytes())
			}
			// Verify that the last 8 bytes correctly encode the int64 decimals.
			decimals := int64(binary.BigEndian.Uint64(data[20:]))
			if decimals != tt.expectedDecimals {
				t.Errorf("decimals mismatch: got %d, want %d", decimals, tt.expectedDecimals)
			}
		})
	}
}

func TestSyncedRewardData_UnmarshalBinary(t *testing.T) {
	// Prepare a valid 28-byte slice for one test case.
	validAddress := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	validDecimals := int64(8)
	validData := make([]byte, 28)
	copy(validData[:20], validAddress.Bytes())
	binary.BigEndian.PutUint64(validData[20:], uint64(validDecimals))

	tests := []struct {
		name      string
		input     []byte
		expected  *syncedRewardData
		expectErr bool
	}{
		{
			name:  "Valid input",
			input: validData,
			expected: &syncedRewardData{
				Erc20Address:  validAddress,
				Erc20Decimals: validDecimals,
			},
			expectErr: false,
		},
		{
			name:      "Invalid input length",
			input:     []byte("short data"),
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sr syncedRewardData
			err := sr.UnmarshalBinary(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected an error for invalid input length, but got none")
				}
				// No need to check further if an error was expected.
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Compare the unmarshaled address.
			if sr.Erc20Address != tt.expected.Erc20Address {
				t.Errorf("address mismatch: got %s, want %s", sr.Erc20Address.Hex(), tt.expected.Erc20Address.Hex())
			}
			// Compare the unmarshaled decimals.
			if sr.Erc20Decimals != tt.expected.Erc20Decimals {
				t.Errorf("decimals mismatch: got %d, want %d", sr.Erc20Decimals, tt.expected.Erc20Decimals)
			}
		})
	}
}
