package crypto_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd25519PrivateKey_Sign(t *testing.T) {
	key := "7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"
	pk, err := crypto.Ed25519PrivateKeyFromHex(key)
	require.NoError(t, err, "error parse private key")

	msg := []byte("foo")

	sig, err := pk.Sign(msg)
	require.NoError(t, err, "error sign")

	expectSignature := "59b2db2d1e4ce6f8771453cfc78d1f943723528f00fa14adf574600f15c601d591fa2ba29c94d9ed694db324f9e8671bdfbcba4b8e10f6a8733682fa3d115f0c"
	assert.Equal(t, expectSignature, hex.EncodeToString(sig), "unexpect signature")
}

func TestEd25519PublicKey_Verify(t *testing.T) {
	key := "0aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"
	keyBytes, err := hex.DecodeString(key)
	require.NoError(t, err, "error decode public key")

	pubKey, err := crypto.Ed25519PublicKeyFromBytes(keyBytes)
	require.NoError(t, err, "error parse public key")

	msg := []byte("foo")
	sig := "59b2db2d1e4ce6f8771453cfc78d1f943723528f00fa14adf574600f15c601d591fa2ba29c94d9ed694db324f9e8671bdfbcba4b8e10f6a8733682fa3d115f0c"
	sigBytes, _ := hex.DecodeString(sig)

	tests := []struct {
		name     string
		msg      []byte
		sigBytes []byte
		wantErr  string // internal error message
	}{
		{
			name:     "verify success",
			msg:      msg,
			sigBytes: sigBytes,
			wantErr:  "",
		},
		{
			name:     "invalid signature length",
			msg:      msg,
			sigBytes: sigBytes[1:],
			wantErr:  "invalid signature",
		},
		{
			name:     "wrong signature",
			msg:      []byte("bar"),
			sigBytes: sigBytes,
			wantErr:  "verify signature failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pubKey.Verify(tt.sigBytes, tt.msg)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err, "expect no error")
			}
		})
	}
}

func TestEd25519PublicKey_Address(t *testing.T) {
	key := "0aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"
	keyBytes, err := hex.DecodeString(key)
	require.NoError(t, err, "error decode public key")

	pubKey, err := crypto.Ed25519PublicKeyFromBytes(keyBytes)
	require.NoError(t, err, "error parse public key")

	eq := pubKey.Address().String() == "0aa611bf555596912bc6f9a9f169f8785918e7ba"
	assert.True(t, eq, "mismatch address")
}

func Test_GenerateEd25518PrivateKey(t *testing.T) {
	pk, err := crypto.GenerateEd25519Key()
	require.NoError(t, err, "error generate key")

	if len(pk.Bytes()) != 64 {
		t.Errorf("invalid private key length: %d", len(pk.Bytes()))
	}
}

func Test_Del(t *testing.T) {
	pkStr := "074e49f310d58d6d9e9060c4d6172cde63565d4227b8f0d4e1106e383134c8f4"
	bts, err := hex.DecodeString(pkStr)
	require.NoError(t, err, "error decode private key")

	pk, err := crypto.Ed25519PublicKeyFromBytes(bts)
	require.NoError(t, err, "error parse private key")

	fmt.Println(pk.Address().String())
	panic("")
}
