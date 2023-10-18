package types

import (
	"fmt"
	"math/big"
)

// TODO: doc it all

type Account struct {
	PublicKey []byte   `json:"public_key"`
	Balance   *big.Int `json:"balance"`
	Nonce     int64    `json:"nonce"`
}

type AccountStatus uint32

const (
	AccountStatusLatest AccountStatus = iota
	AccountStatusPending
)

// ChainInfo describes the current status of a Kwil blockchain.
type ChainInfo struct {
	ChainID     string
	BlockHeight uint64
	BlockHash   string
}

type JoinRequest struct {
	Candidate []byte   // pubkey of the candidate validator
	Power     int64    // the requested power
	ExpiresAt int64    // the block height at which the join request expires
	Board     [][]byte // slice of pubkeys of all the eligible voting validators
	Approved  []bool   // if they approved
}

type Validator struct {
	PubKey []byte
	Power  int64
}

func (v *Validator) String() string {
	return fmt.Sprintf("{pubkey = %x, power = %d}", v.PubKey, v.Power)
}
