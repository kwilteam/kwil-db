package auth_test

import (
	"fmt"
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

func Test_del(t *testing.T) {
	sig := []byte("\xb7Ka\\\xd35LSߨ\x8b\xf6\x9ate*\xf5\r\xd7Q \xd31\xe7\xec}\xd1Wl\t\xe2\x04xv/\x91\xb6f:\xe9\x01Uy)%+V\xa5]\x18#X\x19\xbaa!\x1ek\xe9\xa9\xee>E0\x01")
	sender := []byte("\xaf\xfd\xc0l\xf3J\xfd}X\x01\xa1=H\xc9*Ӗ\t\x90\x1d")
	msg := []byte("https://localhost wants you to sign in with your account:\n\nsign pws\n\nURI: https://localhost/auth\nVersion: 1\nChain ID: kwil-chain-shyc8zBu\nNonce: 236f0c9b3d285df2fe29\nIssue At: 2023-12-05T17:50:13Z\nExpiration Time: 2023-12-05T17:50:43Z\n")

	fmt.Println(sender)
	fmt.Println(string(msg))

	err := auth.EthSecp256k1Authenticator{}.Verify(sender, msg, sig)
	assert.NoError(t, err)

}
