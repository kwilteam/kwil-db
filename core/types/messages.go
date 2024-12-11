package types

// This file contains the messages exchanged between the consensus engine and the block processor.

type BlockExecRequest struct {
	Height   int64
	Block    *Block
	BlockID  Hash
	Proposer []byte
}

type BlockExecResult struct {
	TxResults        []TxResult
	AppHash          Hash
	ValidatorUpdates []*Validator
}

type CommitRequest struct {
	Height  int64
	AppHash Hash
	Syncing bool
}

type ConsensusParams struct {
	// MaxBlockSize is the maximum size of a block in bytes.
	MaxBlockSize int64
	// JoinExpiry is the number of blocks after which the validators
	// join request expires if not approved.
	JoinExpiry int64
	// VoteExpiry is the default number of blocks after which the validators
	// vote expires if not approved.
	VoteExpiry int64
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool

	// MigrationStatus determines the status of the migration.
	MigrationStatus MigrationStatus

	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64
}
