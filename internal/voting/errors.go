package voting

import (
	"errors"
)

var (
	ErrAlreadyProcessed         = errors.New("resolution already processed")
	ErrResolutionAlreadyHasBody = errors.New("resolution already has a body")
)
