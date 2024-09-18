package types

import (
	"fmt"
	"math/big"
)

// TODO: doc it all

type Account struct {
	Identifier HexBytes `json:"identifier"`
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
	GasEnabled  bool   `json:"gas_enabled"`
}

// The validator related types that identify validators by pubkey are still
// []byte, so base64 json marshalling. I'm not sure if they should be hex like
// the account/owner fields in the user service.

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

func (v *Validator) String() string {
	return fmt.Sprintf("{pubkey = %x, power = %d}", v.PubKey, v.Power)
}

// DatasetIdentifier contains the information required to identify a dataset.
type DatasetIdentifier struct {
	Name  string   `json:"name"`
	Owner HexBytes `json:"owner"`
	DBID  string   `json:"dbid"`
}

// VotableEvent is an event that can be voted.
// It contains an event type and a body.
// An ID can be generated from the event type and body.
type VotableEvent struct {
	Type string `json:"type"`
	Body []byte `json:"body"`
}

func (e *VotableEvent) ID() *UUID {
	return NewUUIDV5(append([]byte(e.Type), e.Body...))
}

type PendingResolution struct {
	ResolutionID *UUID    `json:"resolution_id"` // Resolution ID
	ExpiresAt    int64    `json:"expires_at"`    // ExpiresAt is the block height at which the resolution expires
	Board        [][]byte `json:"board"`         // Board is the list of validators who are eligible to vote on the resolution
	Approved     []bool   `json:"approved"`      // Approved is the list of bools indicating if the corresponding validator approved the resolution
}

// Migration is a migration resolution that is proposed by a validator
// for initiating the migration process.
type Migration struct {
	ID               *UUID  `json:"id"`                 // ID is the UUID of the migration resolution/proposal
	ActivationPeriod int64  `json:"activation_height"`  // ActivationPeriod is the amount of blocks before the migration is activated.
	Duration         int64  `json:"migration_duration"` // MigrationDuration is the duration of the migration process starting from the ActivationHeight
	ChainID          string `json:"chain_id"`           // ChainID of the new network
	Timestamp        string `json:"timestamp"`          // Timestamp when the migration was proposed
}

type MigrationStatus int

const (
	NoActiveMigration MigrationStatus = iota
	MigrationNotStarted
	MigrationInProgress
	MigrationCompleted
)

func (status *MigrationStatus) String() string {
	switch *status {
	case NoActiveMigration:
		return "NoActiveMigration"
	case MigrationNotStarted:
		return "MigrationNotStarted"
	case MigrationInProgress:
		return "MigrationInProgress"
	case MigrationCompleted:
		return "MigrationCompleted"
	default:
		return "Unknown"
	}
}

type MigrationState struct {
	Status      MigrationStatus `json:"status"`       // Status is the current status of the migration
	StartHeight int64           `json:"start_height"` // StartHeight is the block height at which the migration started
	EndHeight   int64           `json:"end_height"`   // EndHeight is the block height at which the migration ends
}

// MigrationMetadata holds metadata about a migration, informing
// consumers of what information the current node has available
// for the migration.
type MigrationMetadata struct {
	MigrationState   MigrationState `json:"migration_state"`   // MigrationState is the current state of the migration
	GenesisConfig    []byte         `json:"genesis_config"`    // GenesisConfig is the genesis file data
	SnapshotMetadata []byte         `json:"snapshot_metadata"` // SnapshotMetadata is the snapshot metadata
}
