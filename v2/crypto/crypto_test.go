package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"strings"
	"testing"
)

func TestMarshalUnmarshalPrivateKeyMock(t *testing.T) {
	tests := []struct {
		name    string
		key     PrivateKey
		wantErr bool
	}{
		{
			name:    "nil key",
			key:     nil,
			wantErr: true,
		},
		{
			name:    "empty key bytes",
			key:     &mockPrivateKey{keyType: KeyTypeSecp256k1, bytes: []byte{}},
			wantErr: true,
		},
		{
			name:    "invalid key type",
			key:     &mockPrivateKey{keyType: KeyType(999), bytes: []byte{1, 2, 3}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var marshaled []byte
			if tt.key != nil {
				marshaled = MarshalPrivateKey(tt.key)
			}

			got, err := UnmarshalPrivateKey(marshaled)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !bytes.Equal(got.Bytes(), tt.key.Bytes()) {
				t.Errorf("key bytes mismatch: got %v, want %v", got.Bytes(), tt.key.Bytes())
			}
			if got.Type() != tt.key.Type() {
				t.Errorf("key type mismatch: got %v, want %v", got.Type(), tt.key.Type())
			}
		})
	}
}

func TestUnmarshalPrivateKeyErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr string
	}{
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: "insufficient data for private key",
		},
		{
			name:    "insufficient bytes",
			input:   []byte{1, 2, 3},
			wantErr: "insufficient data for private key",
		},
		{
			name:    "exactly 4 bytes",
			input:   []byte{1, 0, 0, 0},
			wantErr: "insufficient data for private key",
		},
		{
			name:    "unknown key type",
			input:   []byte{255, 255, 255, 255, 1, 2, 3},
			wantErr: "invalid key type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalPrivateKey(tt.input)
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

type mockPrivateKey struct {
	keyType KeyType
	bytes   []byte
	sig     []byte
	sigErr  error
	pubKey  PublicKey
}

func (m *mockPrivateKey) Bytes() []byte {
	return m.bytes
}

func (m *mockPrivateKey) Type() KeyType {
	return m.keyType
}

func (m *mockPrivateKey) Equals(key Key) bool {
	if key == nil {
		return false
	}
	return bytes.Equal(m.bytes, key.Bytes()) && m.keyType == key.Type()
}

func (m *mockPrivateKey) Sign([]byte) ([]byte, error) {
	return m.sig, m.sigErr
}

func (m *mockPrivateKey) Public() PublicKey {
	return m.pubKey
}

type mockPublicKey struct {
	keyType KeyType
	bytes   []byte
	valid   bool
	verErr  error
}

func (m *mockPublicKey) Bytes() []byte {
	return m.bytes
}

func (m *mockPublicKey) Type() KeyType {
	return m.keyType
}

func (m *mockPublicKey) Equal(key Key) bool {
	if key == nil {
		return false
	}
	return bytes.Equal(m.bytes, key.Bytes()) && m.keyType == key.Type()
}

func (m *mockPublicKey) Verify(msg []byte, sig []byte) (bool, error) {
	return m.valid, m.verErr
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
			marshaledKey := MarshalPrivateKey(originalKey)

			// Verify the marshaled data starts with correct key type
			gotKeyType := KeyType(binary.LittleEndian.Uint32(marshaledKey[:4]))
			if gotKeyType != tt.keyType {
				t.Errorf("MarshalPrivateKey() key type = %v, want %v", gotKeyType, tt.keyType)
			}

			// Unmarshal back to private key
			unmarshaledKey, err := UnmarshalPrivateKey(marshaledKey)
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
	t.Run("Invalid key type", func(t *testing.T) {
		invalidKeyType := []byte{255, 255, 255, 255} // Invalid key type
		invalidKey := append(invalidKeyType, make([]byte, 32)...)
		_, err := UnmarshalPrivateKey(invalidKey)
		if err == nil {
			t.Error("Expected error for invalid key type")
		}
	})

	t.Run("Insufficient data", func(t *testing.T) {
		_, err := UnmarshalPrivateKey([]byte{1, 2, 3}) // Less than 4 bytes
		if err == nil {
			t.Error("Expected error for insufficient data")
		}
	})
}
