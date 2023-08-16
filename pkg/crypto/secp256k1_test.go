package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSecp256k1PrivateKey_Sign(t *testing.T) {
	key := "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	pk, err := Secp256k1PrivateKeyFromHex(key)
	require.NoError(t, err, "error parse private key")

	msg := []byte("foo")
	hash := Sha256(msg)

	sig, err := pk.Sign(hash)
	require.NoError(t, err, "error sign")

	expectSig := "19a4aced02d5b9142b4f622b06442b1904445e16bd25409e6b0ff357ccc021d001d0e7824654b695b4b6e0991cb7507f487b82be4b2ed713d1e3e2cbc3d2518d01"
	require.Equal(t, SIGNATURE_SECP256K1_PERSONAL_LENGTH, len(sig), "invalid signature length")
	require.EqualValues(t, hex.EncodeToString(sig), expectSig, "invalid signature")
}

func TestSecp256k1PrivateKey_SignMsg(t *testing.T) {
	key := "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	pk, err := Secp256k1PrivateKeyFromHex(key)
	require.NoError(t, err, "error parse private key")

	msg := []byte("foo")

	sig, err := pk.SignMsg(msg)
	require.NoError(t, err, "error sign msg")

	expectSignature := "cb3fed7f6ff36e59054c04a831b215e514052753ee353e6fe31d4b4ef736acd6155127db555d3006ba14fcb4c79bbad56c8e63b81a9896319bb053a9e253475800"
	expectSignatureBytes, _ := hex.DecodeString(expectSignature)

	expectSig := &Signature{
		Signature: expectSignatureBytes,
		Type:      SIGNATURE_TYPE_SECP256K1_PERSONAL,
	}

	assert.EqualValues(t, expectSig, sig, "unexpect signature")
}

func TestSecp256k1PublicKey_Verify(t *testing.T) {
	key := "04812bef44f6e7b2a19c0b01c2dca5e54ba1935a1890ffdcb93abd0c534b209c21e4f6176823fef493f7b5afaa456f31d5293363d8f801c540ebcc061812890cba"
	keyBytes, err := hex.DecodeString(key)
	require.NoError(t, err, "error decode key")

	pubKey, err := Secp256k1PublicKeyFromBytes(keyBytes)
	require.NoError(t, err, "error parse public key")

	msg := []byte("foo")
	hash := Sha256(msg)

	sig := "19a4aced02d5b9142b4f622b06442b1904445e16bd25409e6b0ff357ccc021d001d0e7824654b695b4b6e0991cb7507f487b82be4b2ed713d1e3e2cbc3d2518d01"
	sigBytes, _ := hex.DecodeString(sig)
	require.Equal(t, SIGNATURE_SECP256K1_PERSONAL_LENGTH, len(sigBytes), "invalid signature length")

	tests := []struct {
		name     string
		sigBytes []byte
		wantErr  error
	}{
		{
			name:     "verify success",
			sigBytes: sigBytes[:len(sigBytes)-1],
			wantErr:  nil,
		},
		{
			name:     "invalid signature length",
			sigBytes: sigBytes,
			wantErr:  errInvalidSignature,
		},
		{
			name:     "wrong signature",
			sigBytes: sigBytes[1:],
			wantErr:  errVerifySignatureFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pubKey.Verify(tt.sigBytes, hash)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSecp256k1PublicKey_Address(t *testing.T) {
	key := "04812bef44f6e7b2a19c0b01c2dca5e54ba1935a1890ffdcb93abd0c534b209c21e4f6176823fef493f7b5afaa456f31d5293363d8f801c540ebcc061812890cba"
	keyBytes, err := hex.DecodeString(key)
	require.NoError(t, err, "error decode key")

	pubKey, err := Secp256k1PublicKeyFromBytes(keyBytes)
	require.NoError(t, err, "error parse public key")

	eq := pubKey.Address().String() == "0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7"
	require.True(t, eq, "mismatch address")
}
