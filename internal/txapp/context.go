package txapp

import "context"

// TxContext is the context for transaction execution.
type TxContext interface {
	// Context gets a Golang context.
	Ctx() context.Context
	// BlockHeight gets the height of the current block.
	BlockHeight() int64
	// Proposer gets the proposer public key of the current block.
	Proposer() []byte
	// ConsensusParams gets the consensus parameters.
	ConsensusParams() ConsensusParams
}

// ConsensusParams contains parameters that are agreed upon by the network.
// These can only be changed via voting on the network.
type ConsensusParams struct {
	// VotingPeriod is the maximum length of a voting period.
	// It is measured in blocks, and is applied additively.
	// e.g. if the current block is 50, and VotingPeriod is 100,
	// then the current voting period ends at block 150.
	VotingPeriod int64
}
