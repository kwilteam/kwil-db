package types

import (
	"fmt"
	"math/big"
)

// TODO: doc it all

type Account struct {
	Identifier []byte   `json:"identifier"`
	Balance    *big.Int `json:"balance"`
	Nonce      int64    `json:"nonce"`
}

type AccountStatus uint32

const (
	AccountStatusLatest AccountStatus = iota
	AccountStatusPending
)

// ChainInfo describes the current status of a Kwil blockchain.
type ChainInfo struct {
	ChainID     string `json:"chain_id"`
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
}

type JoinRequest struct {
	Candidate []byte   `json:"candidate"`  // pubkey of the candidate validator
	Power     int64    `json:"power"`      // the requested power
	ExpiresAt int64    `json:"expires_at"` // the block height at which the join request expires
	Board     [][]byte `json:"board"`      // slice of pubkeys of all the eligible voting validators
	Approved  []bool   `json:"approved"`   // slice of bools indicating if the corresponding validator approved
}

type Validator struct {
	PubKey []byte `json:"pubkey"`
	Power  int64  `json:"power"`
}

// ValidatorRemoveProposal is a proposal from an existing validator (remover) to
// remove a validator (the target) from the validator set.
type ValidatorRemoveProposal struct {
	Target  []byte `json:"target"`  // pubkey of the validator to remove
	Remover []byte `json:"remover"` // pubkey of the validator proposing the removal
}

// NodeInfo contains public information about a node.
// It can be used by clients to join as a peer.
type NodeInfo struct {
	NodeID           string   `json:"node_id"`
	PublicKey        HexBytes `json:"pubkey"`
	P2PListenAddress string   `json:"p2p_listen_address"`
}

func (v *Validator) String() string {
	return fmt.Sprintf("{pubkey = %x, power = %d}", v.PubKey, v.Power)
}
