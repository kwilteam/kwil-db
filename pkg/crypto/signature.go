package crypto

import (
	"fmt"
	"strings"

	ethAccount "github.com/ethereum/go-ethereum/accounts"
)

type SignatureType string

const (
	// SignatureTypeEmpty only used as placeholder
	SignatureTypeEmpty SignatureType = "empty"
	// SignatureTypeInvalid invalid signature type
	SignatureTypeInvalid SignatureType = "invalid"
	//
	SignatureTypeSecp256k1Cometbft SignatureType = "secp256k1_cmt" // secp256k1 cometbft
	SignatureTypeEd25519           SignatureType = "ed25519"
	SignatureTypeSecp256k1Personal SignatureType = "secp256k1_ep" // secp256k1 ethereum personal_sign
)

const (
	SignatureSecp256k1CometbftLength = 64
	SignatureSecp256k1PersonalLength = 65
	SignatureEd25519Length           = 64
)

var SignatureTypeFromName = map[string]SignatureType{
	"secp256k1_cmt": SignatureTypeSecp256k1Cometbft, // secp256k1 cometbft
	"ed25519":       SignatureTypeEd25519,           // ed25519 standard, any better name?
	"secp256k1_ep":  SignatureTypeSecp256k1Personal, // secp256k1 ethereum personal_sign
}

var (
	errInvalidSignature          = fmt.Errorf("invalid signature")
	errVerifySignatureFailed     = fmt.Errorf("verify signature failed")
	errNotSupportedSignatureType = fmt.Errorf("not supported signature type")
)

func SignatureTypeLookUp(name string) SignatureType {
	name = strings.ToLower(name)
	if t, ok := SignatureTypeFromName[name]; ok {
		return t
	}
	return SignatureTypeInvalid
}

func (s SignatureType) KeyType() KeyType {
	switch s {
	case SignatureTypeSecp256k1Cometbft, SignatureTypeSecp256k1Personal:
		return Secp256k1
	case SignatureTypeEd25519:
		return Ed25519
	default:
		return UnknownKeyType
	}
}

func (s SignatureType) String() string {
	return string(s)
}

// Signature is a cryptographic signature.
type Signature struct {
	Signature []byte        `json:"signature_bytes"`
	Type      SignatureType `json:"signature_type"`
}

func (s *Signature) KeyType() KeyType {
	return s.Type.KeyType()
}

// Verify verifies the signature against the given public key and data.
func (s *Signature) Verify(publicKey PublicKey, msg []byte) error {
	switch s.Type {
	case SignatureTypeSecp256k1Personal:
		if len(s.Signature) != SignatureSecp256k1PersonalLength {
			return errInvalidSignature
		}
		hash := ethAccount.TextHash(msg)
		// Remove recovery ID
		sig := s.Signature[:len(s.Signature)-1]
		return publicKey.Verify(sig, hash)
	case SignatureTypeSecp256k1Cometbft:
		if len(s.Signature) != SignatureSecp256k1CometbftLength {
			return errInvalidSignature
		}
		// cometbft using sha256 and 64 bytes signature(no recovery ID 'v')
		hash := Sha256(msg)
		return publicKey.Verify(s.Signature, hash)
	case SignatureTypeEd25519:
		if len(s.Signature) != SignatureEd25519Length {
			return errInvalidSignature
		}
		// hash(sha512) is handled by downstream library
		return publicKey.Verify(s.Signature, msg)
	default:
		return fmt.Errorf("%w: %s", errNotSupportedSignatureType, s.Type.String())
	}
}
