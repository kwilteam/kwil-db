/*
Package auth provides an Authenticator interface for developers to implement
their own Kwil authentication drivers. Authenticator extensions may be used to
expand the type of signatures that may be verified on transactions and messages
via the (*Signature).Verify method, where the Signature.Type field corresponds
to the unique Authenticator name. It also provides the ability to derive an
address from a public key for a certain network.

Similar to Go's database/sql package, developers can implement the Authenticator
interface and register it with the RegisterAuthenticator function. The name used
to register is to be set as the signature type creating signatures.

There are presently two Signers defined in the Kwil Go SDK with pre-registered
Authenticators with the same type: EthPersonalSigType and Ed25519SigType. When
registering a new Authenticator, the values of these may not be used.
*/
package auth

import (
	"errors"
	"fmt"
	"strings"
)

// Authenticator is an interface for authenticating an incoming call
// It is made to work with keypair authentication
type Authenticator interface {
	// Verify verifies the signature against the given public key and data.
	Verify(sender, msg, signature []byte) error

	// Address returns an address from a public key
	Address(sender []byte) (string, error)
}

var registeredAuthenticators = make(map[string]Authenticator)

// RegisterAuthenticator registers an authenticator with a given name
func RegisterAuthenticator(name string, auth Authenticator) error {
	name = strings.ToLower(name)
	if _, ok := registeredAuthenticators[name]; ok {
		return fmt.Errorf("%w: %s", ErrAuthenticatorExists, name)
	}

	registeredAuthenticators[name] = auth
	return nil
}

// getAuthenticator returns an authenticator by the name it was registered with
func getAuthenticator(name string) (Authenticator, error) {
	name = strings.ToLower(name)
	auth, ok := registeredAuthenticators[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, name)
	}

	return auth, nil
}

// GetAddress returns an address from a public key and authenticator type
func GetAddress(authType string, sender []byte) (string, error) {
	auth, err := getAuthenticator(authType)
	if err != nil {
		return "", err
	}

	return auth.Address(sender)
}

var (
	// ErrAuthenticatorExists is returned when an authenticator is already registered
	ErrAuthenticatorExists = errors.New("authenticator already exists")
	// ErrAuthenticatorNotFound is returned when an authenticator is not found
	ErrAuthenticatorNotFound = errors.New("authenticator not found")
)
