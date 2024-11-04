package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestGenerateSecp256k1Key(t *testing.T) {
	tests := []struct {
		name    string
		reader  io.Reader
		wantErr bool
	}{
		{
			name:    "valid random source",
			reader:  bytes.NewReader(bytes.Repeat([]byte{1}, 64)),
			wantErr: false,
		},
		{
			name:    "nil source",
			reader:  nil,
			wantErr: false,
		},
		{
			name:    "empty random source",
			reader:  bytes.NewReader([]byte{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priv, pub, err := GenerateSecp256k1Key(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecp256k1Key() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if priv == nil {
					t.Error("GenerateSecp256k1Key() private key is nil")
				}
				if pub == nil {
					t.Error("GenerateSecp256k1Key() public key is nil")
				}
				if !bytes.Equal(priv.Public().Bytes(), pub.Bytes()) {
					t.Error("GenerateSecp256k1Key() public key mismatch")
				}
			}
		})
	}
}

func TestSecp256k1KeySignVerify(t *testing.T) {
	priv, pub, err := GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("test message")
	sig, err := priv.Sign(msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	valid, err := pub.Verify(msg, sig)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !valid {
		t.Error("Verify() failed for valid signature")
	}

	// Test with modified message
	modifiedMsg := append([]byte{}, msg...)
	modifiedMsg[0] ^= 0xff
	valid, err = pub.Verify(modifiedMsg, sig)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if valid {
		t.Error("Verify() succeeded for modified message")
	}

	// Test with invalid signature format
	invalidSig := []byte("invalid signature format")
	valid, err = pub.Verify(msg, invalidSig)
	if err == nil {
		t.Error("Verify() should fail with invalid signature format")
	}
	if valid {
		t.Error("Verify() succeeded with invalid signature format")
	}
}

func TestUnmarshalSecp256k1Keys(t *testing.T) {
	priv, pub, _ := GenerateSecp256k1Key(rand.Reader)
	privBytes := priv.Bytes()
	pubBytes := pub.Bytes()

	recoveredPub, err := UnmarshalSecp256k1PublicKey(pubBytes)
	if err != nil {
		t.Fatalf("UnmarshalSecp256k1PublicKey() error = %v", err)
	}
	if !pub.Equals(recoveredPub) {
		t.Error("Unmarshaled public key does not match original")
	}

	recoveredPriv, err := UnmarshalSecp256k1PrivateKey(privBytes)
	if err != nil {
		t.Fatalf("UnmarshalSecp256k1PrivateKey() error = %v", err)
	}
	if !priv.Equals(recoveredPriv) {
		t.Error("Unmarshaled private key does not match original")
	}

	// Test invalid key sizes
	_, err = UnmarshalSecp256k1PublicKey(make([]byte, 31))
	if err == nil {
		t.Error("UnmarshalSecp256k1PublicKey() should fail with invalid size")
	}

	_, err = UnmarshalSecp256k1PrivateKey(make([]byte, 31))
	if err == nil {
		t.Error("UnmarshalSecp256k1PrivateKey() should fail with invalid size")
	}
}

func TestSecp256k1KeyEquality(t *testing.T) {
	priv1, pub1, _ := GenerateSecp256k1Key(rand.Reader)
	priv2, pub2, _ := GenerateSecp256k1Key(rand.Reader)

	if priv1.Equals(priv2) {
		t.Error("Different private keys should not be equal")
	}
	if pub1.Equals(pub2) {
		t.Error("Different public keys should not be equal")
	}

	// Test equality with different key types
	edPriv, edPub, _ := GenerateEd25519Key(rand.Reader)
	if priv1.Equals(edPriv) {
		t.Error("Secp256k1 private key should not equal Ed25519 private key")
	}
	if pub1.Equals(edPub) {
		t.Error("Secp256k1 public key should not equal Ed25519 public key")
	}

	mockPriv := &mockPrivateKey{keyType: 99}
	if priv1.Equals(mockPriv) {
		t.Error("Different private key types should not be equal")
	}

	mockPriv = &mockPrivateKey{keyType: KeyTypeSecp256k1} // same KeyType, different bytes
	if priv1.Equals(mockPriv) {
		t.Error("Different bytes should not be equal")
	}

	mockPriv = &mockPrivateKey{
		keyType: KeyTypeSecp256k1,
		bytes:   priv1.Bytes(),
	} // same KeyType, same bytes
	if !priv1.Equals(mockPriv) {
		t.Error("same Type and Bytes should be equal regardless of concrete impl")
	}

	mockPub := &mockPublicKey{keyType: 99}
	if pub1.Equals(mockPub) {
		t.Error("Different private key types should not be equal")
	}

	mockPub = &mockPublicKey{keyType: KeyTypeSecp256k1} // same KeyType, different bytes
	if pub1.Equals(mockPub) {
		t.Error("Different bytes should not be equal")
	}

	mockPub = &mockPublicKey{
		keyType: KeyTypeSecp256k1,
		bytes:   pub1.Bytes(),
	} // same KeyType, same bytes
	if !pub1.Equals(mockPub) {
		t.Error("same Type and Bytes should be equal regardless of concrete impl")
	}
}
