package types

import "context"

type KVStore interface {
	// Set sets a key to a value.
	Set(ctx context.Context, key []byte, value []byte) error
	// Get gets a value for a key.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Delete deletes a key.
	Delete(ctx context.Context, key []byte) error
}
