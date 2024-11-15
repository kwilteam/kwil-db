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

	"kwil/crypto/auth"
)

var (
	// ErrAuthenticatorExists is returned when an authenticator is already registered
	ErrAuthenticatorExists = errors.New("authenticator already exists")
	// ErrAuthenticatorNotFound is returned when an authenticator is not found
	ErrAuthenticatorNotFound = errors.New("authenticator not found")
)

// registeredAuthenticators is the Authenticator registry used by kwild.
var registeredAuthenticators = make(map[string]auth.Authenticator)

// ModOperation is the type used to enumerate authenticator modifications.
type ModOperation int8

// Resolutions may be removed, updated, or added.
const (
	ModRemove ModOperation = iota - 1
	ModUpdate
	ModAdd
)

// RegisterAuthenticator registers, removes, or updates an authenticator with
// the Kwil network.
func RegisterAuthenticator(mod ModOperation, name string, auth auth.Authenticator) error {
	name = strings.ToLower(name)
	if _, ok := registeredAuthenticators[name]; ok {
		switch mod {
		case ModAdd:
			return fmt.Errorf("%w: %s", ErrAuthenticatorExists, name)
		case ModRemove:
			delete(registeredAuthenticators, name)
		case ModUpdate:
			registeredAuthenticators[name] = auth
		}
		return nil
	}
	switch mod {
	case ModRemove, ModUpdate:
		return fmt.Errorf("resolution does not exist to modify (%d)", mod)
	default: // add
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
