package voting

import "errors"

var (
	ErrAlreadyVoted       = errors.New("vote already exists from voter")
	ErrResolutionNotFound = errors.New("resolution not found")
)
