package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
)

// TODO: doc it all

// AccountID uniquely identifies kwil accounts.
// It is a combination of the Identifier and the key type.
// Identifier can be a public key or an address.
type AccountID struct {
	Identifier HexBytes       `json:"identifier"`
	KeyType    crypto.KeyType `json:"key_type"`
}

func (a *AccountID) String() string {
	if a == nil {
		return "AccountID{nil}"
	}

	return fmt.Sprintf("AccountID{identifier = %x, keyType = %s}", a.Identifier, a.KeyType.String())
}

func (id *AccountID) MarshalBinary() ([]byte, error) {
	if id == nil {
		return nil, errors.New("nil account ID")
	}

	buf := new(bytes.Buffer)
	if err := WriteBytes(buf, id.Identifier); err != nil {
		return nil, fmt.Errorf("failed to write account identifier: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, id.KeyType); err != nil {
		return nil, fmt.Errorf("failed to write key type: %w", err)
	}

	return buf.Bytes(), nil
}

func (id *AccountID) UnmarshalBinary(b []byte) error {
	buf := bytes.NewBuffer(b)

	ident, err := ReadBytes(buf)
	if err != nil {
		return fmt.Errorf("failed to read account identifier: %w", err)
	}
	id.Identifier = ident

	if err := binary.Read(buf, binary.LittleEndian, &id.KeyType); err != nil {
		return fmt.Errorf("failed to read key type: %w", err)
	}

	if !id.KeyType.Valid() {
		return fmt.Errorf("invalid key type: %s", id.KeyType)
	}

	return nil
}

// GetSignerAccount returns the account ID of the signer.
func GetSignerAccount(signer auth.Signer) (*AccountID, error) {
	ident := signer.CompactID()

	keyType, err := auth.GetAuthenticatorKeyType(signer.AuthType())
	if err != nil {
		return nil, err
	}

	return &AccountID{
		Identifier: ident,
		KeyType:    keyType,
	}, nil
}

type Account struct {
	ID      *AccountID `json:"id"`
	Balance *big.Int   `json:"balance"`
	Nonce   int64      `json:"nonce"`
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
	BlockHash   Hash   `json:"block_hash"`
}

// The validator related types that identify validators by pubkey are still
// []byte, so base64 json marshalling. I'm not sure if they should be hex like
// the account/owner fields in the user service.

type JoinRequest struct {
	Candidate *AccountID   `json:"candidate"`  // pubkey of the candidate validator
	Power     int64        `json:"power"`      // the requested power
	ExpiresAt int64        `json:"expires_at"` // the block height at which the join request expires
	Board     []*AccountID `json:"board"`      // slice of pubkeys of all the eligible voting validators
	Approved  []bool       `json:"approved"`   // slice of bools indicating if the corresponding validator approved
}

// NodeKey is a public key of a validator node.
type NodeKey struct {
	PubKey HexBytes       `json:"pubkey"`
	Type   crypto.KeyType `json:"type"`
}

func (n *NodeKey) String() string {
	if n == nil {
		return "NodeKey{nil}"
	}
	return fmt.Sprintf("NodeKey{pubkey = %x, keyType = %s}", n.PubKey, n.Type.String())
}

// func (n *NodeKey) AuthType() (string, error) {
// 	switch n.Type {
// 	case crypto.KeyTypeSecp256k1:
// 		return auth.Secp256k1Auth, nil
// 	case crypto.KeyTypeEd25519:
// 		return auth.Ed25519Auth, nil
// 	default:
// 		return "", fmt.Errorf("unknown key type: %s", n.Type)
// 	}
// }

type Validator struct {
	AccountID
	Power int64 `json:"power"`
}

func (v *Validator) String() string {
	return fmt.Sprintf("Validator{pubkey = %x, keyType = %s, power = %d}", v.Identifier, v.KeyType.String(), v.Power)
}

type validatorJSON struct {
	PubKey string `json:"pubkey"`
	Type   string `json:"type"`
	Power  int64  `json:"power"`
}

func (v *Validator) MarshalJSON() ([]byte, error) {
	if v == nil {
		return nil, errors.New("nil Validator")
	}

	return json.Marshal(validatorJSON{
		PubKey: hex.EncodeToString(v.Identifier),
		Type:   v.KeyType.String(),
		Power:  v.Power,
	})

}

func (v *Validator) UnmarshalJSON(b []byte) error {
	var vj validatorJSON
	if err := json.Unmarshal(b, &vj); err != nil {
		return err
	}

	pk, err := hex.DecodeString(vj.PubKey)
	if err != nil {
		return err
	}
	v.Identifier = pk

	kt, err := crypto.ParseKeyType(vj.Type)
	if err != nil {
		return err
	}
	v.KeyType = kt

	v.Power = vj.Power
	return nil
}

// DatasetIdentifier contains the information required to identify a dataset.
type DatasetIdentifier struct {
	Name  string   `json:"name"`
	Owner HexBytes `json:"owner"`
	DBID  string   `json:"dbid"`
}

// VotableEventID returns the ID of an event that can be voted on. This may be
// used to determine the ID of an event prior to the event being created.
func VotableEventID(ty string, body []byte) UUID {
	return *NewUUIDV5(append([]byte(ty), body...))
}

// VotableEvent is an event that can be voted.
// It contains an event type and a body.
// An ID can be generated from the event type and body.
type VotableEvent struct {
	Type string `json:"type"`
	Body []byte `json:"body"`
}

func (e *VotableEvent) ID() *UUID {
	id := VotableEventID(e.Type, e.Body)
	return &id
}

const veVersion = 0

func (e VotableEvent) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	// version uint16 first
	if err := binary.Write(buf, binary.BigEndian, uint16(veVersion)); err != nil {
		return nil, err
	}
	WriteString(buf, e.Type)
	WriteBytes(buf, e.Body) // could also buf.Write(e.Body) since this is the last field
	return buf.Bytes(), nil
}

func (e *VotableEvent) UnmarshalBinary(b []byte) error {
	buf := bytes.NewBuffer(b)
	var version uint16
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return err
	}
	if version != veVersion {
		return fmt.Errorf("invalid version: %d", version)
	}
	eType, err := ReadString(buf)
	if err != nil {
		return err
	}
	eBody, err := ReadBytes(buf)
	if err != nil {
		return err
	}
	e.Type = eType
	e.Body = eBody
	return nil
}

type PendingResolution struct {
	Type         string       `json:"type"`
	ResolutionID *UUID        `json:"resolution_id"` // Resolution ID
	ExpiresAt    int64        `json:"expires_at"`    // ExpiresAt is the block height at which the resolution expires
	Board        []*AccountID `json:"board"`         // Board is the list of validators who are eligible to vote on the resolution
	Approved     []bool       `json:"approved"`      // Approved is the list of bools indicating if the corresponding validator approved the resolution
}

// Migration is a migration resolution that is proposed by a validator
// for initiating the migration process.
type Migration struct {
	ID               *UUID  `json:"id"`                 // ID is the UUID of the migration resolution/proposal
	ActivationPeriod int64  `json:"activation_height"`  // ActivationPeriod is the amount of blocks before the migration is activated.
	Duration         int64  `json:"migration_duration"` // MigrationDuration is the duration of the migration process starting from the ActivationHeight
	Timestamp        string `json:"timestamp"`          // Timestamp when the migration was proposed
}

type ConsensusParamUpdateProposal struct {
	ID          UUID         `json:"id"`
	Description string       `json:"description"`
	Updates     ParamUpdates `json:"updates"`
}

// MigrationStatus represents the status of the nodes in the zero downtime migration process.
type MigrationStatus string

func (ms MigrationStatus) NoneActive() bool {
	return ms == NoActiveMigration || ms == ""
}

func (ms MigrationStatus) Active() bool {
	return !ms.NoneActive()
}

func (ms MigrationStatus) Valid() bool {
	switch ms {
	case NoActiveMigration, ActivationPeriod, MigrationInProgress,
		MigrationCompleted, GenesisMigration:
		return true
	default:
		return false
	}
}

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
	Validators []*Validator `json:"validators"`
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

	BlockTimestamp int64 `json:"block_time"` // epoch millis
	BlockAge       int64 `json:"block_age"`  // milliseconds
	Syncing        bool  `json:"syncing"`
	Height         int64 `json:"height"`
	AppHash        Hash  `json:"app_hash"`
	PeerCount      int   `json:"peer_count"`

	// Mode is an oddball field as it pertains to the service config rather than
	// state of the node. It is provided here as a convenience so applications
	// can discern node state and the mode of interaction with one request.
	Mode ServiceMode `json:"mode"` // e.g. "private"
}
