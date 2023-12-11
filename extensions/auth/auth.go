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
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	// internal/ident is home to the Authenticator registry used by kwild, as
	// well as the registry-powered verification functions used by kwild. The
	// RegisterAuthenticator helper is provided here so that extension
	// implementations may register themselves on import, but it would be fine
	// to shift that responsibility to the importing code in kwild (these stubs)
)

var (
	// ErrAuthenticatorExists is returned when an authenticator is already registered
	ErrAuthenticatorExists = errors.New("authenticator already exists")
	// ErrAuthenticatorNotFound is returned when an authenticator is not found
	ErrAuthenticatorNotFound = errors.New("authenticator not found")
)

// registeredAuthenticators is the Authenticator registry used by kwild.
var registeredAuthenticators = make(map[string]auth.Authenticator)

// RegisterAuthenticator registers an authenticator with a given name
func RegisterAuthenticator(name string, auth auth.Authenticator) error {
	name = strings.ToLower(name)
	if _, ok := registeredAuthenticators[name]; ok {
		return fmt.Errorf("%w: %s", ErrAuthenticatorExists, name)
	}

	registeredAuthenticators[name] = auth
	return nil
}

// getAuthenticator returns an authenticator by the name it was registered with
func GetAuthenticator(name string) (auth.Authenticator, error) {
	name = strings.ToLower(name)
	auth, ok := registeredAuthenticators[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, name)
	}

	return auth, nil
}
