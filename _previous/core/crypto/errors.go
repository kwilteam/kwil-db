package crypto

import (
	"errors"
)

var (
	ErrInvalidSignature       = errors.New("signature verification failed")
	ErrInvalidSignatureLength = errors.New("invalid signature length")
)
