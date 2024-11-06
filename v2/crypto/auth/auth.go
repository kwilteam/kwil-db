/*
Package auth provides the standard signing and verification methods used in
Kwil. These are Ethereum "personal sign" used by wallets to sign a customized
readable message, and plain Ed25519 signing used for validator node signatures.

It also defines an Authenticator interface for developers to implement their own
Kwil authentication drivers. See the extensions/auth package in the kwil-db main
module. Authenticator extensions may be used to expand the type of signatures
that may be verified on transactions and messages. It also provides the ability
to derive an address from a public key for a certain network.

There are presently two Signers defined in the Kwil Go SDK with pre-registered
Authenticators with the same type: EthPersonalSigType and Ed25519SigType. When
registering a new Authenticator, the values of these may not be used. This is
the primary reason that the Authenticator interface is defined in this package
instead of the kwil-db main module under extensions/auth. We may consider moving
these two Authenticator implementations out of the SDK and into the main module
where they are only available to the application that needs them, but it may be
awkward to have complementary verification defined in the same place as the
signing.
*/
package auth

// Authenticator is an interface for verifying signatures and
// deriving a string identifier from the sender bytes. Custom
// asymmetric signature algorithms may be implemented by developers
// by implementing this interface.
type Authenticator interface {
	// Verify verifies whether a signature is valid for a given
	// message and sender. It is meant to be used with asymmetric
	// signature algorithms such as ECDSA, Ed25519 RSA, etc. If the
	// signature is invalid, the method should return an error. If
	// the signature is valid, the method should return nil.
	Verify(sender, msg, signature []byte) error

	// Identifier returns a string identifier for a given sender.
	// This string identifier is used to identify the sender when
	// interacting with the Kuneiform engine, and will be used as
	// the `@caller` variable in the engine.
	Identifier(sender []byte) (string, error)
}

func GetAuthenticator(authType string) Authenticator {
	switch authType {
	case EthPersonalSignAuth:
		return EthSecp256k1Authenticator{}
	case Ed25519Auth:
		return Ed25519Authenticator{}
	default:
		return nil
	}
}
