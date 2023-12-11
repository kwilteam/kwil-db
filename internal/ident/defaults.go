package ident

import (
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	authExt "github.com/kwilteam/kwil-db/extensions/auth"
)

// Register the two Authenticators required by kwild. The implementations of
// these are defined in the SDK (core module) since their counterpart signers
// must correspond exactly in their message handling.

func init() {
	err := authExt.RegisterAuthenticator(auth.Ed25519Auth, auth.Ed25519Authenticator{})
	if err != nil {
		panic(err)
	}

	err = authExt.RegisterAuthenticator(auth.EthPersonalSignAuth, auth.EthSecp256k1Authenticator{})
	if err != nil {
		panic(err)
	}
}
