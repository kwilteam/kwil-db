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
	PubKey HexBytes `json:"pubkey"`
	Power  int64    `json:"power"`
}

// ValidatorRemoveProposal is a proposal from an existing validator (remover) to
// remove a validator (the target) from the validator set.
type ValidatorRemoveProposal struct {
	Target  []byte `json:"target"`  // pubkey of the validator to remove
	Remover []byte `json:"remover"` // pubkey of the validator proposing the removal
}

func (v *Validator) String() string {
	return fmt.Sprintf("Validator{pubkey = %x, power = %d}", v.PubKey, v.Power)
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
	Timestamp        string `json:"timestamp"`          // Timestamp when the migration was proposed
}

// MigrationStatus represents the status of the nodes in the zero downtime migration process.
type MigrationStatus string

const (
	// NoActiveMigration indicates there is currently no migration process happening on the network.
	NoActiveMigration MigrationStatus = "NoActiveMigration"

	// ActivationPeriod represents the phase after the migration proposal has been approved by the network,
	// but before the migration begins. During this phase, validators prepare their nodes for migration.
	ActivationPeriod MigrationStatus = "ActivationPeriod"

	// MigrationInProgress indicates that the nodes on the old network are in migration mode and
	// records the state changes to be replicated on the new network.
	MigrationInProgress MigrationStatus = "MigrationInProgress"

	// MigrationCompleted indicates that the migration process has successfully finished on the old network,
	// and the old network is ready to be decommissioned once the new network has caught up.
	MigrationCompleted MigrationStatus = "MigrationCompleted"

	// GenesisMigration refers to the phase where the nodes on the new network during migration bootstraps
	// with the genesis state and replicates the state changes from the old network.
	GenesisMigration MigrationStatus = "GenesisMigration"
)

type MigrationState struct {
	Status        MigrationStatus `json:"status"`                 // Status is the current status of the migration
	StartHeight   int64           `json:"start_height,omitempty"` // StartHeight is the block height at which the migration started on the old chain
	EndHeight     int64           `json:"end_height,omitempty"`   // EndHeight is the block height at which the migration ends on the old chain
	CurrentHeight int64           `json:"chain_height,omitempty"` // CurrentHeight is the current block height of the node
}

// MigrationMetadata holds metadata about a migration, informing
// consumers of what information the current node has available
// for the migration.
type MigrationMetadata struct {
	MigrationState   MigrationState `json:"migration_state"`   // MigrationState is the current state of the migration
	GenesisInfo      *GenesisInfo   `json:"genesis_info"`      // GenesisInfo is the genesis information
	SnapshotMetadata []byte         `json:"snapshot_metadata"` // SnapshotMetadata is the snapshot metadata
	Version          int            `json:"version"`           // Version of the migration metadata
}

// GenesisInfo holds the genesis information that the new network should use
// in its genesis file.
type GenesisInfo struct {
	// AppHash is the application hash of the old network at the StartHeight
	AppHash HexBytes `json:"app_hash"`
	// Validators is the list of validators that the new network should start with
	Validators []*NamedValidator `json:"validators"`
}

// NamedValidator is a validator with a name.
// Since CometBFT assigns validators human-readable names, this struct
// is used to represent a validator with its name that will be used in the genesis file.
type NamedValidator struct {
	Name      string `json:"name"`
	Validator `json:"validator"`
}

// ServiceMode describes the operating mode of the user service. Namely, if the
// service is in private mode (where calls are authenticated, query is disabled,
// and raw transactions cannot be retrieved).
type ServiceMode string

const (
	ModeOpen    ServiceMode = "open"
	ModePrivate ServiceMode = "private"
)

// Health is the response for MethodHealth. This determines the
// serialized response for the Health method required by the rpcserver.Svc
// interface. This is the response with which most health checks will be concerned.
type Health struct {
	ChainInfo

	// Healthy is based on several factors determined by the service and it's
	// configuration, such as the maximum age of the best block and if the node
	// is still syncing (in catch-up or replay).
	Healthy bool `json:"healthy"`

	// Version is the service API version.
	Version string `json:"version"`

	BlockTimestamp int64    `json:"block_time"` // epoch millis
	BlockAge       int64    `json:"block_age"`  // milliseconds
	Syncing        bool     `json:"syncing"`
	AppHeight      int64    `json:"app_height"` // may be less than block store best block
	AppHash        HexBytes `json:"app_hash"`
	PeerCount      int      `json:"peer_count"`

	// Mode is an oddball field as it pertains to the service config rather than
	// state of the node. It is provided here as a convenience so applications
	// can discern node state and the mode of interaction with one request.
	Mode ServiceMode `json:"mode"` // e.g. "private"
}
