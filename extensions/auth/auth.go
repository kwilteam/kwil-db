/*
Package auth contains any known Authenticator extensions that may be selected at
build-time for use in kwild. Authenticator extensions are used to expand the
type of signatures that may be verified, and define address derivation for the
public keys of the corresponding type.

Build constraints a.k.a. build tags are used to enable extensions in a kwild
binary. See README.md in the extensions package for more information.
*/
package auth

import (
	"github.com/kwilteam/kwil-db/core/crypto/auth"

	// internal/ident is home to the Authenticator registry used by kwild, as
	// well as the registry-powered verification functions used by kwild. The
	// RegisterAuthenticator helper is provided here so that extension
	// implementations may register themselves on import, but it would be fine
	// to shift that responsibility to the importing code in kwild (these stubs)
	"github.com/kwilteam/kwil-db/internal/ident"
)

func RegisterAuthenticator(name string, auth auth.Authenticator) error {
	return ident.RegisterAuthenticator(name, auth)
}

// var RegisterAuthenticator = ident.RegisterAuthenticator
