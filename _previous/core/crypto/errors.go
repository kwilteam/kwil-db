package crypto

import (
	"errors"
)

var (
	ErrInvalidSignature       = errors.New("invalid signature")
	ErrInvalidSignatureLength = errors.New("invalid signature length")
)
