package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"strings"
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

func TestVitalik(t *testing.T) {
	// sig and sig hash for mainnet eth tx
	// 0x1190dd5cd7f1bed506aa7a76fe060d2fc6a0214543b701e01b6748eb6ed16196
	var rawSig [65]byte
	rB, _ := hex.DecodeString("8a1c54556e2aaaf86ade107060b55df9e7d53651158958f91b4c377d012894d6")
	sB, _ := hex.DecodeString("5cd64506f16258f6d3831b3445cb94a0de70ff8943abb4a39a0a84ec1bb81d6d")
	copy(rawSig[:], rB)
	copy(rawSig[32:], sB)
	rawSig[64] = 1 // v

	// var r, s secp256k1.ModNScalar
	// if r.SetByteSlice(rawSig[:32]) {
	// 	t.Fatal("r value overflow")
	// }
	// if s.SetByteSlice(rawSig[32:]) {
	// 	t.Fatal("s value overflow")
	// }
	// sig := ecdsa.NewSignature(&r, &s)

	hash, _ := hex.DecodeString("a752f7d86cee4952d18b0976b84f1532f59a0ebc534e12ae1f99b75def8cff1d")

	pub, err := RecoverSecp256k1Key(hash[:], rawSig[:])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%x", pub.SerializeUncompressed())
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

func TestComputeEthereumAddress(t *testing.T) {
	tests := []struct {
		name     string
		pubKey   func() *Secp256k1PublicKey
		wantAddr string
	}{
		{
			name: "expected key and address",
			pubKey: func() *Secp256k1PublicKey {
				privKey, _, _ := GenerateSecp256k1Key(bytes.NewReader(bytes.Repeat([]byte{1}, 64)))
				return privKey.Public().(*Secp256k1PublicKey)
			},
			wantAddr: "1a642f0e3c3af545e7acbd38b07251b3990914f1",
		},
		{
			name: "publicly known",
			pubKey: func() *Secp256k1PublicKey {
				pubkeyhex, _ := hex.DecodeString("0462117d6727ddd50b8f1d60ce50ef9fa511c7b43b6b6e6f763b32b942e515a4d47df6eb61d3dceb615176c80a16484e773885f3de31e0344ed3d74cce103646f4")
				pubkey, err := UnmarshalSecp256k1PublicKey(pubkeyhex)
				if err != nil {
					t.Fatal(err)
				}
				return pubkey
			},
			wantAddr: "9cea81b9d2e900d6027125378ee2ddfa15feeed1",
		},
		{
			name: "vitalik",
			pubKey: func() *Secp256k1PublicKey {
				pubkeyhex, _ := hex.DecodeString("04e95ba0b752d75197a8bad8d2e6ed4b9eb60a1e8b08d257927d0df4f3ea6860992aac5e614a83f1ebe4019300373591268da38871df019f694f8e3190e493e711")
				pubkey, err := UnmarshalSecp256k1PublicKey(pubkeyhex)
				if err != nil {
					t.Fatal(err)
				}
				return pubkey
			},
			wantAddr: strings.ToLower("d8dA6BF26964aF9D7eEd9e03E53415D37aA96045"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey := tt.pubKey()
			addr := EthereumAddressFromPubKey(pubKey)

			if tt.wantAddr != "" {
				wantBytes, err := hex.DecodeString(tt.wantAddr)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(addr, wantBytes) {
					t.Errorf("ComputeEthereumAddress() = %x, want %s", addr, tt.wantAddr)
				}
			} else {
				if len(addr) != 20 {
					t.Errorf("ComputeEthereumAddress() returned address of length %d, want 20", len(addr))
				}
			}

			// Verify address is always 20 bytes
			if len(addr) != 20 {
				t.Errorf("ComputeEthereumAddress() returned address of incorrect length: got %d, want 20", len(addr))
			}

			// Verify generating address twice for same key returns same result
			addr2 := EthereumAddressFromPubKey(pubKey)
			if !bytes.Equal(addr, addr2) {
				t.Error("ComputeEthereumAddress() returned different addresses for same key")
			}
		})
	}
}
