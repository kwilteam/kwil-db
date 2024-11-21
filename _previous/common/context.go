package common

// TODO: review these contexts and see if all the fields are necessary
// ChainContext provides context for all chain operations.
// Fields in ChainContext should never be mutated, except
// NetworkParameters can be deterministically mutated as part
// of block execution.
type ChainContext struct {
	// ChainID is the unique identifier for the chain.
	ChainID string

	// NetworkParams holds network level parameters that can be evolved
	// over the lifetime of a network.
	NetworkParameters *NetworkParameters

	// MigrationParams holds the context for all migration operations such as
	// block info to poll for the changesets from the old chain during migration.
	MigrationParams *MigrationContext
}

// BlockContext provides context for all block operations.
type BlockContext struct {
	// ChainContext contains information about the chain.
	ChainContext *ChainContext
	// Height gets the height of the current block.
	Height int64
	// Timestamp is a timestamp of the current block.
	// It is set by the block proposer, and therefore may not be accurate.
	// It should not be used for time-sensitive operations where incorrect
	// timestamps could result in security vulnerabilities.
	Timestamp int64
	// Proposer gets the proposer public key of the current block.
	Proposer []byte
}

// MigrationContext provides context for all migration operations.
// Fields in MigrationContext should never be mutated till the migration is completed.
type MigrationContext struct {
	// StartHeight is the height of the first block to start migration.
	StartHeight int64
	// EndHeight is the height of the last block to end migration.
	EndHeight int64
}
