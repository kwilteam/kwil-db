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

// GetAuthenticator returns an authenticator by the name it was registered with
func GetAuthenticator(name string) (Authenticator, error) {
	name = strings.ToLower(name)
	auth, ok := registeredAuthenticators[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrAuthenticatorNotFound, name)
	}

	return auth, nil
}

// ListAuthenticators returns a list of registered authenticators
func ListAuthenticators() []struct {
	Name          string
	Authenticator Authenticator
} {
	var authenticators []struct {
		Name          string
		Authenticator Authenticator
	}

	for name, auth := range registeredAuthenticators {
		authenticators = append(authenticators, struct {
			Name          string
			Authenticator Authenticator
		}{
			Name:          name,
			Authenticator: auth,
		})
	}

	return authenticators
}

var (
	// ErrAuthenticatorExists is returned when an authenticator is already registered
	ErrAuthenticatorExists = errors.New("authenticator already exists")
	// ErrAuthenticatorNotFound is returned when an authenticator is not found
	ErrAuthenticatorNotFound = errors.New("authenticator not found")
)
