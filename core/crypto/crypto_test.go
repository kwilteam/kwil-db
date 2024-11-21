package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
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
				marshaled = WireEncodePrivateKey(tt.key)
			}

			got, err := WireDecodePrivateKey(marshaled)
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
			_, err := WireDecodePrivateKey(tt.input)
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

var _ PrivateKey = (*mockPrivateKey)(nil)

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

var _ PublicKey = (*mockPublicKey)(nil)

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

func (m *mockPublicKey) Equals(key Key) bool {
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
			marshaledKey := WireEncodePrivateKey(originalKey)

			// Verify the marshaled data starts with correct key type
			gotKeyType := KeyType(binary.LittleEndian.Uint32(marshaledKey[:4]))
			if gotKeyType != tt.keyType {
				t.Errorf("MarshalPrivateKey() key type = %v, want %v", gotKeyType, tt.keyType)
			}

			// Unmarshal back to private key
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
	t.Run("Invalid key type", func(t *testing.T) {
		invalidKeyType := []byte{255, 255, 255, 255} // Invalid key type
		invalidKey := append(invalidKeyType, make([]byte, 32)...)
		_, err := WireDecodePrivateKey(invalidKey)
		if err == nil {
			t.Error("Expected error for invalid key type")
		}
	})

	t.Run("Insufficient data", func(t *testing.T) {
		_, err := WireDecodePrivateKey([]byte{1, 2, 3}) // Less than 4 bytes
		if err == nil {
			t.Error("Expected error for insufficient data")
		}
	})
}

func TestKeyEquals(t *testing.T) {
	secp256k1Key1, _, _ := GenerateSecp256k1Key(rand.Reader)
	secp256k1Key2, _, _ := GenerateSecp256k1Key(rand.Reader)
	ed25519Key1, _, _ := GenerateEd25519Key(rand.Reader)
	ed25519Key2, _, _ := GenerateEd25519Key(rand.Reader)
	mockKey1 := &mockPrivateKey{keyType: KeyTypeSecp256k1, bytes: []byte{1, 2, 3}}
	mockKey2 := &mockPrivateKey{keyType: KeyTypeSecp256k1, bytes: []byte{1, 2, 3}}
	mockKey3 := &mockPrivateKey{keyType: KeyTypeSecp256k1, bytes: []byte{4, 5, 6}}

	tests := []struct {
		name string
		k1   Key
		k2   Key
		want bool
	}{
		{
			name: "same interface value",
			k1:   secp256k1Key1,
			k2:   secp256k1Key1,
			want: true,
		},
		{
			name: "different secp256k1 private keys",
			k1:   secp256k1Key1,
			k2:   secp256k1Key2,
			want: false,
		},
		{
			name: "secp256k1 private key and its public key",
			k1:   secp256k1Key1,
			k2:   secp256k1Key1.Public(),
			want: false,
		},
		{
			name: "different ed25519 private keys",
			k1:   ed25519Key1,
			k2:   ed25519Key2,
			want: false,
		},
		{
			name: "ed25519 private key and its public key",
			k1:   ed25519Key1,
			k2:   ed25519Key1.Public(),
			want: false,
		},
		{
			name: "identical mock keys",
			k1:   mockKey1,
			k2:   mockKey2,
			want: true,
		},
		{
			name: "different mock keys",
			k1:   mockKey1,
			k2:   mockKey3,
			want: false,
		},
		{
			name: "secp256k1 public keys from same private key",
			k1:   secp256k1Key1.Public(),
			k2:   secp256k1Key1.Public(),
			want: true,
		},
		{
			name: "ed25519 public keys from same private key",
			k1:   ed25519Key1.Public(),
			k2:   ed25519Key1.Public(),
			want: true,
		},
		{
			name: "nil keys",
			k1:   nil,
			k2:   nil,
			want: true,
		},
		{
			name: "one nil key",
			k1:   secp256k1Key1,
			k2:   nil,
			want: false,
		},
		{
			name: "other nil key",
			k1:   nil,
			k2:   secp256k1Key1,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KeyEquals(tt.k1, tt.k2); got != tt.want {
				t.Errorf("KeyEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandlePanic(t *testing.T) {
	originalPanicWriter := panicWriter
	defer func() {
		panicWriter = originalPanicWriter
	}()

	tests := []struct {
		name            string
		rerr            interface{}
		where           string
		wantErr         bool
		wantErrContains string
		setupWriter     func() *bytes.Buffer
	}{
		{
			name:    "nil panic",
			rerr:    nil,
			where:   "test",
			wantErr: false,
			setupWriter: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
		{
			name:            "string panic",
			rerr:            "test panic",
			where:           "testFunc",
			wantErr:         true,
			wantErrContains: "panic in testFunc: test panic",
			setupWriter: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
		{
			name:            "error panic",
			rerr:            errors.New("error panic"),
			where:           "errorTest",
			wantErr:         true,
			wantErrContains: "panic in errorTest: error panic",
			setupWriter: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
		{
			name:            "integer panic",
			rerr:            42,
			where:           "intTest",
			wantErr:         true,
			wantErrContains: "panic in intTest: 42",
			setupWriter: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
		{
			name:            "struct panic",
			rerr:            struct{ msg string }{"structured panic"},
			where:           "structTest",
			wantErr:         true,
			wantErrContains: "panic in structTest: {structured panic}",
			setupWriter: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := tt.setupWriter()
			panicWriter = buf

			var err error
			handlePanic(tt.rerr, &err, tt.where)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("error = %v, want containing %v", err, tt.wantErrContains)
				}
				if !strings.Contains(buf.String(), "caught panic") {
					t.Error("expected panic message in output buffer")
				}
				if !strings.Contains(buf.String(), "goroutine") {
					t.Error("expected stack trace in output buffer")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if buf.Len() > 0 {
					t.Error("expected empty output buffer")
				}
			}
		})
	}
}
