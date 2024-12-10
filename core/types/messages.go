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
