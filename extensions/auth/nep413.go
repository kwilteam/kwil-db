//go:build auth_nep413 || ext_test

package auth

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	borsch "github.com/near/borsh-go"
)

func init() {
	err := RegisterAuthenticator(ModAdd, nep413Name, Nep413Authenticator{
		MsgEncoder: func(bts []byte) string {
			return string(bts)
		},
	})
	if err != nil {
		panic(err)
	}
}

const nep413Name = "nep413"

/*

	This implements the NEP-413 standard as a Kwil authentication driver.

	NEP-413 is a standard for authentication developed by NEAR protocol. It utilizes
	borsch serialization and ed25519 signatures to authenticate users.

	More info can be found here: https://github.com/near/NEPs/blob/master/neps/nep-0413.md.

	In order to implement this, the incoming `Signature` field needs to contain more data that just the ed25519 signature.
	It needs to contain the following:
	- the length of the first struct (the nep413 `Payload`, without the Tag and Message)
	- the first struct (the nep413 `Payload`, without the Tag and Message), serialized with borsch
	- the signature generated from the nep413 `AuthenticationToken.signature`

	The first struct is the nep413 `Payload` (https://github.com/near/NEPs/blob/master/neps/nep-0413.md#structure).
	All fields of the payload EXCEPT the message and tag are included here (the message field
	is populated by the serialized Kwil transaction, and the tag is standard in nep413).
	Below, the first struct is identified by the struct type `nep413Payload`.

	The signature is the `signature` field returned from the nep413 signer function (https://github.com/near/NEPs/blob/master/neps/nep-0413.md#output-interface).
	This does not contain the the other fields, because they are not needed.

	The signature bytes are as follows:
		1. [0-1]: uint16 length of the first struct (big endian)
		2. [2-2+length]: payload struct
		3. [2+length:]: signature
*/

// Nep413Payload is the message sent to the NEP-413 signer.
// It utilizes borsch for deterministic serialization
type Nep413Payload struct {
	// Tag is some NEAR specific thing that is not really explained anywhere,
	// but should always be the number 2^31+413, or 2147484061
	// https://github.com/near/NEPs/blob/master/neps/nep-0413.md#example
	Tag uint32

	// Message is the plaintext message
	Message string

	// Nonce is the 32 byte nonce of the message
	Nonce [32]byte

	// Recipient is the string identifier of the recipient (e.g. satoshi.near)
	Recipient string

	// CallbackUrl is the url to call when the signature is ready
	CallbackUrl *string
}

type Nep413Authenticator struct {
	// MsgEncoder is a function that encodes a message into a string.
	// For production, this should be base64.StdEncoding.EncodeToString,
	// but for testing, it can be a different encoding.
	MsgEncoder func([]byte) string
}

func (n Nep413Authenticator) Verify(sender []byte, msg []byte, signature []byte) error {
	if len(sender) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid ed25519 public key size when verifying signature %d", len(sender))
	}

	// deserialize the signature
	payload, sig, err := deserializeSignature(signature)
	if err != nil {
		return err
	}

	// add the tag and message to the payload
	payload.Tag = 2147484061
	payload.Message = n.MsgEncoder(msg)

	// serialize the payload
	// it is critical that the payload is not a pointer,
	// since it gives a different result.
	payloadBytes, err := borsch.Serialize(*payload)
	if err != nil {
		return err
	}

	// hash the payload
	hash := sha256.Sum256(payloadBytes)

	// verify the signature
	if !ed25519.Verify(sender, hash[:], sig) {
		return errors.New("signature verification failed")
	}

	return nil
}

// Identifier generates a NEAR implicit address from a public key,
// which is simply the hex-encoded public key.
func (n Nep413Authenticator) Identifier(sender []byte) (string, error) {
	if len(sender) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid ed25519 public key size for generating near address: %d", len(sender))
	}

	return hex.EncodeToString(sender), nil
}

// deserializeSignature deserializes the payload and the signature from the signature bytes
func deserializeSignature(signatureBts []byte) (*Nep413Payload, []byte, error) {
	if len(signatureBts) < 2 {
		return nil, nil, errors.New("invalid signature length")
	}

	// get the length of the first struct
	payloadLength := int(signatureBts[0])<<8 | int(signatureBts[1]) // big endian

	// check that the length is valid
	if len(signatureBts) < 2+payloadLength {
		return nil, nil, errors.New("invalid signature length")
	}

	struct1Bts := signatureBts[2 : 2+payloadLength]

	// deserialize the first struct
	payload := &Nep413Payload{}
	err := borsch.Deserialize(payload, struct1Bts)
	if err != nil {
		return nil, nil, err
	}

	// borsch-go has a weird bug where it will deserialize a nil string ((*string)(nil))
	// as an empty string (""). When serializing, it does NOT
	// serialize an empty string as a nil string. This accounts for that.
	if payload.CallbackUrl != nil && *payload.CallbackUrl == "" {
		payload.CallbackUrl = nil
	}

	signature := signatureBts[2+payloadLength:]
	if len(signature) != ed25519.SignatureSize {
		return nil, nil, errors.New("invalid ed25519 signature length")
	}

	return payload, signature, nil
}
