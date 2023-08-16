package crypto

import (
	"fmt"
	ethAccount "github.com/ethereum/go-ethereum/accounts"
)

type SignatureType int32

const (
	SIGNATURE_TYPE_INVALID SignatureType = iota
	SIGNATURE_TYPE_EMPTY
	SIGNATURE_TYPE_SECP256K1_COMETBFT
	SIGNATURE_TYPE_SECP256K1_PERSONAL // ethereum EIP-191 personal_sign
	SIGNATURE_TYPE_ED25519
	END_SIGNATURE_TYPE
)

const (
	SIGNATURE_SECP256K1_COMETBFT_LENGTH = 64
	SIGNATURE_SECP256K1_PERSONAL_LENGTH = 65
	SIGNATURE_ED25519_LENGTH            = 64
)

var (
	errInvalidSignature          = fmt.Errorf("invalid signature")
	errVerifySignatureFailed     = fmt.Errorf("verify signature failed")
	errNotSupportedSignatureType = fmt.Errorf("not supported signature type")
)

// IsValid returns an error if the signature type is invalid.
func (s SignatureType) IsValid() error {
	if s < SIGNATURE_TYPE_INVALID || s >= END_SIGNATURE_TYPE {
		return fmt.Errorf("%w: %d", errNotSupportedSignatureType, s)
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
		panic("not supported signature type")
	}
}

// Signature is a cryptographic signature.
type Signature struct {
	Signature []byte        `json:"signature_bytes"`
	Type      SignatureType `json:"signature_type"`
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
