package crypto

import (
	"fmt"
)

type SignatureType int32

const (
	SIGNATURE_TYPE_INVALID SignatureType = iota
	SIGNATURE_TYPE_EMPTY
	SIGNATURE_TYPE_ED25519
	END_SIGNATURE_TYPE
)

// IsValid returns an error if the signature type is invalid.
func (s *SignatureType) IsValid() error {
	if *s < SIGNATURE_TYPE_INVALID || *s >= END_SIGNATURE_TYPE {
		return fmt.Errorf("invalid signature type '%d'", *s)
	}
	return nil
}

// Int32 returns the signature type as an int32.
func (s SignatureType) Int32() int32 {
	return int32(s)
}

// Signature is a cryptographic signature.
type Signature struct {
	Signature []byte        `json:"signature_bytes"`
	Type      SignatureType `json:"signature_type"`
}

// Verify verifies the signature against the given public key and data.
func (s *Signature) Verify(publicKey PublicKey, data []byte) error {
	return publicKey.Verify(s, data)
}
