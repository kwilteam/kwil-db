package auth_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
)

const (
	secp256k1Key = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	ed25519Key   = "7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"
)

func Test_Auth(t *testing.T) {

	// testCase will take a signer
	// it will sign a message and verify the signature using
	// the proper authenticator.  It will then check that the
	// address is correct
	type testCase struct {
		name          string
		signer        auth.Signer
		authenticator auth.Authenticator
		address       string
	}

	var msg = []byte("foo")

	testCases := []testCase{
		{
			name:          "eth personal sign",
			signer:        newEthSigner(secp256k1Key),
			authenticator: auth.EthSecp256k1Authenticator{},
			address:       "0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7",
		},
		{
			name:          "ed25519",
			signer:        newEd25519Signer(ed25519Key),
			authenticator: auth.Ed25519Authenticator{},
			// ed25519 doesn't really have the concept of address, so it is just the hex public key
			address: "0aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sig, err := tc.signer.Sign(msg)
			assert.NoError(t, err)

			// verify the signature
			err = tc.authenticator.Verify(tc.signer.Identity(), msg, sig.Signature)
			assert.NoError(t, err)

			// check the address
			address, err := tc.authenticator.Identifier(tc.signer.Identity())
			assert.NoError(t, err)
			assert.Equal(t, tc.address, address)
		})
	}
}

func newEthSigner(pkey string) *auth.EthPersonalSigner {
	secpKey, err := crypto.Secp256k1PrivateKeyFromHex(pkey)
	if err != nil {
		panic(err)
	}

	return &auth.EthPersonalSigner{Key: *secpKey}
}

func newEd25519Signer(pkey string) *auth.Ed25519Signer {
	edKey, err := crypto.Ed25519PrivateKeyFromHex(pkey)
	if err != nil {
		panic(err)
	}

	return &auth.Ed25519Signer{Ed25519PrivateKey: *edKey}
}
