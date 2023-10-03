package crypto_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: all the tests below are using same key pair and message

func TestSecp256k1PrivateKey_Sign(t *testing.T) {
	key := "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	pk, err := crypto.Secp256k1PrivateKeyFromHex(key)
	require.NoError(t, err, "error parse private key")

	msg := []byte("foo")
	hash := sha256.Sum256(msg)

	sig, err := pk.SignWithRecoveryID(hash[:])
	require.NoError(t, err, "error sign")

	expectSig := "19a4aced02d5b9142b4f622b06442b1904445e16bd25409e6b0ff357ccc021d001d0e7824654b695b4b6e0991cb7507f487b82be4b2ed713d1e3e2cbc3d2518d01"
	require.Equal(t, 65, len(sig), "invalid signature length")
	require.EqualValues(t, hex.EncodeToString(sig), expectSig, "invalid signature")
}

func TestSecp256k1PublicKey_Verify(t *testing.T) {
	key := "04812bef44f6e7b2a19c0b01c2dca5e54ba1935a1890ffdcb93abd0c534b209c21e4f6176823fef493f7b5afaa456f31d5293363d8f801c540ebcc061812890cba"
	keyBytes, err := hex.DecodeString(key)
	require.NoError(t, err, "error decode key")

	pubKey, err := crypto.Secp256k1PublicKeyFromBytes(keyBytes)
	require.NoError(t, err, "error parse public key")

	msg := []byte("foo")
	hash := sha256.Sum256(msg)

	sig := "19a4aced02d5b9142b4f622b06442b1904445e16bd25409e6b0ff357ccc021d001d0e7824654b695b4b6e0991cb7507f487b82be4b2ed713d1e3e2cbc3d2518d01"
	sigBytes, _ := hex.DecodeString(sig)
	require.Equal(t, 65, len(sigBytes), "invalid signature length")

	tests := []struct {
		name     string
		sigBytes []byte
		wantErr  error
	}{
		{
			name:     "verify success with 65 bytes signature",
			sigBytes: sigBytes[:],
			wantErr:  nil,
		},
		{
			name:     "verify success with 64 bytes signature(no recovery ID 'v')",
			sigBytes: sigBytes[:len(sigBytes)-1],
			wantErr:  nil,
		},
		{
			name:     "wrong signature",
			sigBytes: sigBytes[1:],
			wantErr:  crypto.ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pubKey.Verify(tt.sigBytes, hash[:])
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "verify error")
				return
			}

			assert.NoError(t, err, "verify error")
		})
	}
}
