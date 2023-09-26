// Package validators provides a federated network validator manager that
// persists the validator set and computes validator updates as transactions are
// processed and blocks finalized. In this system, nodes may request to join the
// validator set, and the existing nodes approve (or not) the request. The
// request passes if 2/3 of the validator set at the time of the request
// approves. A validator is identified by their public key. A node's power is
// intended weight their approvals, and to indicate if they are to be removed.
// Removal is typically initiated from a leave transaction or the consensus
// engine punishing a validator for some bad behavior.
package validators

import (
	"fmt"
)

type Validator struct {
	PubKey []byte
	Power  int64
}

func (v *Validator) String() string {
	return fmt.Sprintf("{pubkey = %x, power = %d}", v.PubKey, v.Power)
}

type JoinRequest struct {
	Candidate []byte   // pubkey of the candidate validator
	Power     int64    // the requested power
	ExpiresAt int64    // the block height at which the join request expires
	Board     [][]byte // slice of pubkeys of all the eligible voting validators
	Approved  []bool   // if they approved
}
