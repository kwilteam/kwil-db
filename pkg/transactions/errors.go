package transactions

import "errors"

var (
	ErrFailedHashReconstruction = errors.New("failed to reconstruct hash")
)
