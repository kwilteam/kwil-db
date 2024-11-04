package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestGenerateEd25519Key(t *testing.T) {
	tests := []struct {
		name    string
		reader  *bytes.Reader
		wantErr bool
	}{
		{
			name:    "valid random source",
			reader:  bytes.NewReader(bytes.Repeat([]byte{1}, 64)),
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
			priv, pub, err := GenerateEd25519Key(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEd25519Key() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if priv == nil {
					t.Error("GenerateEd25519Key() private key is nil")
				}
				if pub == nil {
					t.Error("GenerateEd25519Key() public key is nil")
				}
				if !bytes.Equal(priv.Public().Bytes(), pub.Bytes()) {
					t.Error("GenerateEd25519Key() public key mismatch")
				}
			}
		})
	}
}

func TestEd25519KeySignVerify(t *testing.T) {
	priv, pub, err := GenerateEd25519Key(rand.Reader)
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

	// Test invalid signature
	invalidSig := make([]byte, len(sig))
	copy(invalidSig, sig)
	invalidSig[0] ^= 0xff
	valid, err = pub.Verify(msg, invalidSig)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if valid {
		t.Error("Verify() succeeded for invalid signature")
	}
}

func TestEd25519KeyEquality(t *testing.T) {
	priv1, pub1, _ := GenerateEd25519Key(rand.Reader)
	priv2, pub2, _ := GenerateEd25519Key(rand.Reader)

	if priv1.Equals(priv2) {
		t.Error("Different private keys should not be equal")
	}
	if pub1.Equals(pub2) {
		t.Error("Different public keys should not be equal")
	}
	if !priv1.Public().Equals(pub1) {
		t.Error("Derived public key should equal original public key")
	}
}

func TestUnmarshalEd25519Keys(t *testing.T) {
	priv, pub, _ := GenerateEd25519Key(rand.Reader)
	privBytes := priv.Bytes()
	pubBytes := pub.Bytes()

	recoveredPub, err := UnmarshalEd25519PublicKey(pubBytes)
	if err != nil {
		t.Fatalf("UnmarshalEd25519PublicKey() error = %v", err)
	}
	if !pub.Equals(recoveredPub) {
		t.Error("Unmarshaled public key does not match original")
	}

	recoveredPriv, err := UnmarshalEd25519PrivateKey(privBytes)
	if err != nil {
		t.Fatalf("UnmarshalEd25519PrivateKey() error = %v", err)
	}
	if !priv.Equals(recoveredPriv) {
		t.Error("Unmarshaled private key does not match original")
	}

	// Test invalid key sizes
	_, err = UnmarshalEd25519PublicKey(make([]byte, 31))
	if err == nil {
		t.Error("UnmarshalEd25519PublicKey() should fail with invalid size")
	}

	_, err = UnmarshalEd25519PrivateKey(make([]byte, 63))
	if err == nil {
		t.Error("UnmarshalEd25519PrivateKey() should fail with invalid size")
	}
}
