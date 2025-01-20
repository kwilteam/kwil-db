package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
)

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestMarshalUnmarshalPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		genKey  func() PrivateKey
		keyType KeyType
	}{
		{
			name: "Secp256k1",
			genKey: func() PrivateKey {
				key, _, _ := GenerateSecp256k1Key(rand.Reader)
				return key
			},
			keyType: KeyTypeSecp256k1,
		},
		{
			name: "Ed25519",
			genKey: func() PrivateKey {
				key, _, _ := GenerateEd25519Key(rand.Reader)
				return key
			},
			keyType: KeyTypeEd25519,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a new private key
			originalKey := tt.genKey()

			// Marshal the key
			marshaledKey := WireEncodeKey(originalKey)

			// Verify the marshaled data starts with correct key type
			gotKeyType, _, err := WireDecodeKeyType(marshaledKey)
			if err != nil {
				t.Fatalf("WireDecodeKeyType() error = %v", err)
			}
			// gotKeyType := KeyType(binary.LittleEndian.Uint32(marshaledKey[:4]))
			if gotKeyType != tt.keyType {
				t.Errorf("MarshalPrivateKey() key type = %v, want %v", gotKeyType, tt.keyType)
			}

			unmarshaledKey, err := WireDecodePrivateKey(marshaledKey)
			if err != nil {
				t.Fatalf("UnmarshalPrivateKey() error = %v", err)
			}

			// Verify the unmarshaled key matches original
			if !KeyEquals(originalKey, unmarshaledKey) {
				t.Errorf("Unmarshaled key does not match original key")
			}

			// Test signing and verification works with unmarshaled key
			msg := []byte("test message")
			sig, err := unmarshaledKey.Sign(msg)
			if err != nil {
				t.Fatalf("Sign() error = %v", err)
			}

			valid, err := unmarshaledKey.Public().Verify(msg, sig)
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if !valid {
				t.Error("Signature verification failed")
			}
		})
	}

	// Test error cases
	t.Run("bad key type", func(t *testing.T) {
		_, err := WireDecodePrivateKey([]byte{255, 255, 255, 255})
		if err == nil {
			t.Error("Expected error for insufficient data")
		}
	})

	t.Run("Insufficient data", func(t *testing.T) {
		_, _, err := WireDecodeKeyType([]byte{1, 2, 3}) // Less than 4 bytes
		if err == nil {
			t.Error("Expected error for insufficient data")
		}
	})
}

func TestParseKeyType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    KeyType
		wantErr bool
	}{
		{
			name:    "valid secp256k1",
			input:   "secp256k1",
			want:    KeyTypeSecp256k1,
			wantErr: false,
		},
		{
			name:    "valid ed25519",
			input:   "ed25519",
			want:    KeyTypeEd25519,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "mixed case secp256k1",
			input:   "SECP256K1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "mixed case ed25519",
			input:   "Ed25519",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid key type",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "numeric string",
			input:   "123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "special characters",
			input:   "secp256k1!",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeyType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKeyType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseKeyType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalKeyTypeErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr string
	}{
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: "invalid key type encoding",
		},
		{
			name:    "insufficient bytes",
			input:   []byte{1, 2, 3},
			wantErr: "invalid key type encoding",
		},
		{
			name:    "unknown key type",
			input:   []byte{0, 0, 0, 255},
			wantErr: "unknown key type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := WireDecodeKeyType(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestUnmarshalPublicKey(t *testing.T) {
	secK := mustDecodeHex("02ec1139961d355bf456e659c895a0baee7d214e9229516ee657bca48539971b33")
	edK := mustDecodeHex("5301bc297f9125b6e9ecaea84256bd681da165e53d8fb52effbcda15afa8ee41")
	tests := []struct {
		name    string
		data    []byte
		keyType KeyType
		wantErr bool
	}{
		{
			name:    "valid ed25519",
			data:    edK,
			keyType: KeyTypeEd25519,
			wantErr: false,
		},
		{
			name:    "valid secp256k1",
			data:    secK,
			keyType: KeyTypeSecp256k1,
			wantErr: false,
		},
		{
			name:    "empty data secp256k1",
			data:    []byte{},
			keyType: KeyTypeSecp256k1,
			wantErr: true,
		},
		{
			name:    "empty data ed25519",
			data:    []byte{},
			keyType: KeyTypeEd25519,
			wantErr: true,
		},
		{
			name:    "invalid key type",
			data:    []byte{1, 2, 3},
			keyType: KeyType("invalid"),
			wantErr: true,
		},
		{
			name:    "corrupted secp256k1 data",
			data:    []byte{1, 2, 3, 4, 5},
			keyType: KeyTypeSecp256k1,
			wantErr: true,
		},
		{
			name:    "corrupted ed25519 data",
			data:    []byte{1, 2, 3, 4, 5},
			keyType: KeyTypeEd25519,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalPublicKey(tt.data, tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("UnmarshalPublicKey() returned nil without error")
			}
		})
	}
}

func TestUnmarshalPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		keyType KeyType
		wantErr bool
	}{
		{
			name:    "empty data secp256k1",
			data:    []byte{},
			keyType: KeyTypeSecp256k1,
			wantErr: true,
		},
		{
			name:    "empty data ed25519",
			data:    []byte{},
			keyType: KeyTypeEd25519,
			wantErr: true,
		},
		{
			name:    "invalid key type",
			data:    []byte{1, 2, 3},
			keyType: KeyType("invalid"),
			wantErr: true,
		},
		{
			name:    "corrupted secp256k1 data",
			data:    []byte{1, 2, 3, 4, 5},
			keyType: KeyTypeSecp256k1,
			wantErr: true,
		},
		{
			name:    "corrupted ed25519 data",
			data:    []byte{1, 2, 3, 4, 5},
			keyType: KeyTypeEd25519,
			wantErr: true,
		},
		{
			name:    "invalid length secp256k1",
			data:    make([]byte, 31),
			keyType: KeyTypeSecp256k1,
			wantErr: true,
		},
		{
			name:    "invalid length ed25519",
			data:    make([]byte, 63),
			keyType: KeyTypeEd25519,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalPrivateKey(tt.data, tt.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("UnmarshalPrivateKey() returned nil without error")
			}
		})
	}
}

func TestWireEncodeDecodePublicKey(t *testing.T) {
	tests := []struct {
		name    string
		genKey  func() PublicKey
		keyType KeyType
		mutate  func([]byte) []byte
		wantErr bool
	}{
		{
			name: "encode decode secp256k1 public key",
			genKey: func() PublicKey {
				_, pub, _ := GenerateSecp256k1Key(rand.Reader)
				return pub
			},
			keyType: KeyTypeSecp256k1,
			wantErr: false,
		},
		{
			name: "encode decode ed25519 public key",
			genKey: func() PublicKey {
				_, pub, _ := GenerateEd25519Key(rand.Reader)
				return pub
			},
			keyType: KeyTypeEd25519,
			wantErr: false,
		},
		{
			name: "truncated key bytes",
			genKey: func() PublicKey {
				_, pub, _ := GenerateSecp256k1Key(rand.Reader)
				return pub
			},
			mutate: func(b []byte) []byte {
				return b[:len(b)-1]
			},
			wantErr: true,
		},
		{
			name: "extra key bytes",
			genKey: func() PublicKey {
				_, pub, _ := GenerateSecp256k1Key(rand.Reader)
				return pub
			},
			mutate: func(b []byte) []byte {
				return append(b, 0x00)
			},
			wantErr: true,
		},
		{
			name: "corrupted key type",
			genKey: func() PublicKey {
				_, pub, _ := GenerateSecp256k1Key(rand.Reader)
				return pub
			},
			mutate: func(b []byte) []byte {
				b[0] = 0xFF
				b[1] = 0xFF
				b[2] = 0xFF
				b[3] = 0xFF
				return b
			},
			wantErr: true,
		},
		{
			name: "exactly 4 bytes",
			genKey: func() PublicKey {
				_, pub, _ := GenerateSecp256k1Key(rand.Reader)
				return pub
			},
			mutate: func(b []byte) []byte {
				return b[:4]
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalKey := tt.genKey()
			encoded := WireEncodeKey(originalKey)

			if tt.mutate != nil {
				encoded = tt.mutate(encoded)
			}

			decoded, err := WireDecodePubKey(encoded)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !KeyEquals(originalKey, decoded) {
				t.Error("decoded key does not match original")
			}

			if decoded.Type() != originalKey.Type() {
				t.Errorf("key type mismatch: got %v, want %v", decoded.Type(), originalKey.Type())
			}

			if !bytes.Equal(decoded.Bytes(), originalKey.Bytes()) {
				t.Error("key bytes do not match")
			}
		})
	}
}
