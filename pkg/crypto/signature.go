package crypto

import (
	"fmt"
	ethAccount "github.com/ethereum/go-ethereum/accounts"
	"strings"
)

type SignatureType int32

const (
	SIGNATURE_TYPE_INVALID SignatureType = iota
	SIGNATURE_TYPE_EMPTY
	SIGNATURE_TYPE_SECP256K1_COMETBFT
	SIGNATURE_TYPE_ED25519
	SIGNATURE_TYPE_SECP256K1_PERSONAL // ethereum EIP-191 personal_sign
	END_SIGNATURE_TYPE
)

const (
	SIGNATURE_SECP256K1_COMETBFT_LENGTH = 64
	SIGNATURE_SECP256K1_PERSONAL_LENGTH = 65
	SIGNATURE_ED25519_LENGTH            = 64
)

var SignatureTypeNames = [...]string{
	"invalid",
	"empty",
	"secp256k1_ct",
	"ed25519",
	"secp256k1_ep",
	"invalid",
}

var SignatureTypeFromName = map[string]SignatureType{
	"secp256k1_ct": SIGNATURE_TYPE_SECP256K1_COMETBFT, // secp256k1 cometbft
	"ed25519":      SIGNATURE_TYPE_ED25519,            // ed25519 standard
	"secp256k1_ep": SIGNATURE_TYPE_SECP256K1_PERSONAL, // secp256k1 ethereum personal_sign
}

var (
	errInvalidSignature          = fmt.Errorf("invalid signature")
	errVerifySignatureFailed     = fmt.Errorf("verify signature failed")
	errNotSupportedSignatureType = fmt.Errorf("not supported signature type")
)

func SignatureLookUp(name string) SignatureType {
	name = strings.ToLower(name)
	if t, ok := SignatureTypeFromName[name]; ok {
		return t
	}
	return SIGNATURE_TYPE_INVALID
}

// IsValid returns an error if the signature type is invalid.
func (s SignatureType) IsValid() error {
	if s <= SIGNATURE_TYPE_INVALID || s >= END_SIGNATURE_TYPE {
		return fmt.Errorf("%w: %s", errNotSupportedSignatureType, s.String())
	}
	return nil
}

// Int32 returns the signature type as an int32.
func (s SignatureType) Int32() int32 {
	return int32(s)
}

func (s SignatureType) KeyType() KeyType {
	switch s {
	case SIGNATURE_TYPE_SECP256K1_COMETBFT, SIGNATURE_TYPE_SECP256K1_PERSONAL:
		return Secp256k1
	case SIGNATURE_TYPE_ED25519:
		return Ed25519
	default:
		return UnknownKeyType
	}
}

func (s SignatureType) String() string {
	if s <= SIGNATURE_TYPE_INVALID || s >= END_SIGNATURE_TYPE {
		return "invalid"
	}
	return SignatureTypeNames[s]
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
	case SIGNATURE_TYPE_SECP256K1_PERSONAL:
		if len(s.Signature) != SIGNATURE_SECP256K1_PERSONAL_LENGTH {
			return errInvalidSignature
		}
		hash := ethAccount.TextHash(msg)
		// Remove recovery ID
		sig := s.Signature[:len(s.Signature)-1]
		return publicKey.Verify(sig, hash)
	case SIGNATURE_TYPE_SECP256K1_COMETBFT:
		if len(s.Signature) != SIGNATURE_SECP256K1_COMETBFT_LENGTH {
			return errInvalidSignature
		}
		// cometbft using sha256 and 64 bytes signature(no recovery ID 'v')
		hash := Sha256(msg)
		return publicKey.Verify(s.Signature, hash)
	case SIGNATURE_TYPE_ED25519:
		if len(s.Signature) != SIGNATURE_ED25519_LENGTH {
			return errInvalidSignature
		}
		// hash(sha512) is handled by downstream library
		return publicKey.Verify(s.Signature, msg)
	default:
		return fmt.Errorf("%w: %d", errNotSupportedSignatureType, s.Type)
	}
}
