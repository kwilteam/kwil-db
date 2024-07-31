package common

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
}

// BlockContext provides context for all block operations.
type BlockContext struct {
	// ChainContext contains information about the chain.
	ChainContext *ChainContext
	// Height gets the height of the current block.
	Height int64
	// BlockTimestamp is a timestamp of the current block.
	// It is set by the block proposer, and therefore may not be accurate.
	// It should not be used for time-sensitive operations where incorrect
	// timestamps could result in security vulnerabilities.
	BlockTimestamp int64
	// Proposer gets the proposer public key of the current block.
	Proposer []byte
}
