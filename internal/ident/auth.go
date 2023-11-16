/*
Package ident provides the functions required by kwild for message and
transaction signature verification, and address derivation. Out of the box it
supports verification with signatures created by one of the required default
signers in core/crypto/auth.

It also contains the Authenticator registry used by these methods. The registry
is populated by: (1) automatically registering the default Authenticators, and
(2) importing any auth extensions at or near the level of package main.
*/
package ident

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/crypto/auth" // Signature type and Authenticator interface
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
func getAuthenticator(name string) (auth.Authenticator, error) {
	name = strings.ToLower(name)
	auth, ok := registeredAuthenticators[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, name)
	}

	return auth, nil
}

// verifySig verifies the signature given a signer's public key and the message.
// The type of the Signature determines how the message digest is prepared, and
// what key type is used. The function requires an Authenticator to be
// registered for the signature type. See VerifyTransaction and VerifyMessage in
// verify.go.
func verifySig(pubkey, msg []byte, sig *auth.Signature) error {
	authn, err := getAuthenticator(sig.Type)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, sig.Type)
	}
	return authn.Verify(pubkey, msg, sig.Signature)
}

// Identifier returns a string identifier from a sender and authenticator type. The
// function requires an Authenticator to be registered for the signature type.
func Identifier(authType string, sender []byte) (string, error) {
	authn, err := getAuthenticator(authType)
	if err != nil {
		return "", err
	}
	return authn.Identifier(sender)
}
