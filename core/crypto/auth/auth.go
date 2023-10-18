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

// Verifier is satisfied by types that can verify a signature against a public
// key and message. A Verifier implementation will generally pertain to a
// certain message serialization scheme and key type.
type Verifier interface {
	// Verify verifies the signature against the given public key and data.
	Verify(sender, msg, signature []byte) error
}

// Authenticator is an interface for authenticating a message and deriving an
// encoded address for a public key.
type Authenticator interface {
	Verifier

	// Address returns an address from a public key
	Address(sender []byte) (string, error)
}
