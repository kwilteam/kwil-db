package addresses

import (
	"errors"
)

// errors
var (
	ErrInvalidKeyType      = errors.New("invalid key type")
	ErrIncompatibleAddress = errors.New("specified address format is incompatible with the key type")
)
