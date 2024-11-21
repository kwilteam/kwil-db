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

	"github.com/kwilteam/kwil-db/core/crypto/auth"        // Signature type and Authenticator interface
	authExt "github.com/kwilteam/kwil-db/extensions/auth" // Authenticator registry
)

var (
	// ErrAuthenticatorExists is returned when an authenticator is already registered
	ErrAuthenticatorExists = errors.New("authenticator already exists")
	// ErrAuthenticatorNotFound is returned when an authenticator is not found
	ErrAuthenticatorNotFound = errors.New("authenticator not found")
)

// verifySig verifies the signature given a signer's identity and the message.
// The type of the Signature determines how the message digest is prepared, and
// what key type is used. The function requires an Authenticator to be
// registered for the signature type. See VerifyTransaction and VerifyMessage in
// verify.go.
func verifySig(identity, msg []byte, sig *auth.Signature) error {
	authn, err := authExt.GetAuthenticator(sig.Type)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, sig.Type)
	}
	return authn.Verify(identity, msg, sig.Signature)
}

// Identifier returns a string identifier from a sender and authenticator type. The
// function requires an Authenticator to be registered for the signature type.
func Identifier(authType string, sender []byte) (string, error) {
	authn, err := authExt.GetAuthenticator(authType)
	if err != nil {
		return "", err
	}
	return authn.Identifier(sender)
}
