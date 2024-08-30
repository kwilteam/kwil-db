package transactions

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

// PayloadType is the type of payload
type PayloadType string

func (p PayloadType) String() string {
	return string(p)
}

// Native payload types
const (
	PayloadTypeDeploySchema        PayloadType = "deploy_schema"
	PayloadTypeDropSchema          PayloadType = "drop_schema"
	PayloadTypeExecute             PayloadType = "execute"
	PayloadTypeTransfer            PayloadType = "transfer"
	PayloadTypeValidatorJoin       PayloadType = "validator_join"
	PayloadTypeValidatorLeave      PayloadType = "validator_leave"
	PayloadTypeValidatorRemove     PayloadType = "validator_remove"
	PayloadTypeValidatorApprove    PayloadType = "validator_approve"
	PayloadTypeValidatorVoteIDs    PayloadType = "validator_vote_ids"
	PayloadTypeValidatorVoteBodies PayloadType = "validator_vote_bodies"
	PayloadTypeCreateResolution    PayloadType = "create_resolution"
	PayloadTypeApproveResolution   PayloadType = "approve_resolution"
	PayloadTypeDeleteResolution    PayloadType = "delete_resolution"
)

// payloadConcreteTypes associates a payload type with the concrete type of
// Payload. Use with UnmarshalPayload or reflect to instantiate.
var payloadConcreteTypes = map[PayloadType]Payload{
	PayloadTypeDropSchema:          &DropSchema{},
	PayloadTypeDeploySchema:        &Schema{},
	PayloadTypeExecute:             &ActionExecution{},
	PayloadTypeValidatorJoin:       &ValidatorJoin{},
	PayloadTypeValidatorApprove:    &ValidatorApprove{},
	PayloadTypeValidatorRemove:     &ValidatorRemove{},
	PayloadTypeValidatorLeave:      &ValidatorLeave{},
	PayloadTypeTransfer:            &Transfer{},
	PayloadTypeValidatorVoteIDs:    &ValidatorVoteIDs{},
	PayloadTypeValidatorVoteBodies: &ValidatorVoteBodies{},
	PayloadTypeCreateResolution:    &CreateResolution{},
	PayloadTypeApproveResolution:   &ApproveResolution{},
	PayloadTypeDeleteResolution:    &DeleteResolution{},
}

// UnmarshalPayload unmarshals a serialized transaction payload into an instance
// of the type registered for the given PayloadType.
func UnmarshalPayload(payloadType PayloadType, payload []byte) (Payload, error) {
	prototype, have := payloadConcreteTypes[payloadType]
	if !have {
		return nil, errors.New("unknown payload type")
	}

	t := reflect.TypeOf(prototype).Elem()
	elem := reflect.New(t)       // reflect.Type => reflect.Value
	instance := elem.Interface() // reflect.Type => any

	err := serialize.Decode(payload, instance)
	if err != nil {
		return nil, err
	}
	payloadIface, ok := instance.(Payload)
	if !ok { // should be impossible since payloadConcreteTypes maps to a Payload
		return nil, errors.New("instance not a payload")
	}
	return payloadIface, nil
}

// Valid says if the payload type is known. This does not mean that the node
// will execute the transaction, e.g. not yet activated, or removed.
func (p PayloadType) Valid() bool {
	// native types first for speed
	switch p {
	case PayloadTypeDeploySchema,
		PayloadTypeDropSchema,
		PayloadTypeExecute,
		PayloadTypeValidatorJoin,
		PayloadTypeValidatorApprove,
		PayloadTypeValidatorRemove,
		PayloadTypeValidatorLeave,
		PayloadTypeTransfer,
		PayloadTypeCreateResolution,
		PayloadTypeApproveResolution,
		PayloadTypeDeleteResolution,
		// These should not come in user transactions, but they are not invalid
		// payload types in general.
		PayloadTypeValidatorVoteIDs,
		PayloadTypeValidatorVoteBodies:

		return true
	default: // check map that includes registered payloads from extensions
		return payloadTypes[p]
	}
}

// payloadTypes includes native types and types registered from extensions.
var payloadTypes = map[PayloadType]bool{
	PayloadTypeDeploySchema:        true,
	PayloadTypeDropSchema:          true,
	PayloadTypeExecute:             true,
	PayloadTypeTransfer:            true,
	PayloadTypeValidatorJoin:       true,
	PayloadTypeValidatorLeave:      true,
	PayloadTypeValidatorRemove:     true,
	PayloadTypeValidatorApprove:    true,
	PayloadTypeValidatorVoteIDs:    true,
	PayloadTypeValidatorVoteBodies: true,
	PayloadTypeCreateResolution:    true,
	PayloadTypeApproveResolution:   true,
	PayloadTypeDeleteResolution:    true,
}

// RegisterPayload registers a new payload type. This should be done on
// application initialization. A known payload type does not require a
// corresponding route handler to be registered with extensions/consensus so
// that they become available for consensus according to chain config.
func RegisterPayload(pType PayloadType) {
	if _, have := payloadTypes[pType]; have {
		panic(fmt.Sprintf("already have payload type %v", pType))
	}
	payloadTypes[pType] = true
}

// Payload is the interface that all payloads must implement
// Implementations should use Kwil's serialization package to encode and decode themselves
type Payload interface {
	MarshalBinary() (serialize.SerializedData, error)
	UnmarshalBinary(serialize.SerializedData) error
	Type() PayloadType
}

var _ Payload = (*Schema)(nil)

// DropSchema is the payload that is used to drop a schema
type DropSchema struct {
	DBID string
}

var _ Payload = (*DropSchema)(nil)

func (s *DropSchema) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(s)
}

func (s *DropSchema) UnmarshalBinary(b serialize.SerializedData) error {
	return serialize.Decode(b, s)
}

func (s *DropSchema) Type() PayloadType {
	return PayloadTypeDropSchema
}

// ActionExecution is the payload that is used to execute an action
type ActionExecution struct {
	DBID      string
	Action    string
	Arguments [][]*types.EncodedValue
}

var _ Payload = (*ActionExecution)(nil)

func (a *ActionExecution) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(a)
}

func (a *ActionExecution) UnmarshalBinary(b serialize.SerializedData) error {
	return serialize.Decode(b, a)
}

func (a *ActionExecution) Type() PayloadType {
	return PayloadTypeExecute
}

// Transfer transfers an amount of tokens from the sender to the receiver.
type Transfer struct {
	To     []byte `json:"to"`     // to be string as user identifier
	Amount string `json:"amount"` // big.Int
}

func (v *Transfer) Type() PayloadType {
	return PayloadTypeTransfer
}

var _ encoding.BinaryUnmarshaler = (*Transfer)(nil)
var _ encoding.BinaryMarshaler = (*Transfer)(nil)

func (v *Transfer) UnmarshalBinary(b []byte) error {
	return serialize.Decode(b, v)
}

func (v *Transfer) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

// ValidatorJoin requests to join the network with
// a certain amount of power
type ValidatorJoin struct {
	Power uint64
}

func (v *ValidatorJoin) Type() PayloadType {
	return PayloadTypeValidatorJoin
}

var _ encoding.BinaryUnmarshaler = (*ValidatorJoin)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorJoin)(nil)

func (v *ValidatorJoin) UnmarshalBinary(b []byte) error {
	return serialize.Decode(b, v)
}

func (v *ValidatorJoin) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

// ValidatorApprove is used to vote for a validators approval to join the network
type ValidatorApprove struct {
	Candidate []byte
}

func (v *ValidatorApprove) Type() PayloadType {
	return PayloadTypeValidatorApprove
}

var _ encoding.BinaryUnmarshaler = (*ValidatorApprove)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorApprove)(nil)

func (v *ValidatorApprove) UnmarshalBinary(b []byte) error {
	return serialize.Decode(b, v)
}

func (v *ValidatorApprove) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

// ValidatorRemove is used to vote for a validators removal from the network
type ValidatorRemove struct {
	Validator []byte
}

func (v *ValidatorRemove) Type() PayloadType {
	return PayloadTypeValidatorRemove
}

var _ encoding.BinaryUnmarshaler = (*ValidatorRemove)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorRemove)(nil)

func (v *ValidatorRemove) UnmarshalBinary(b []byte) error {
	return serialize.Decode(b, v)
}

func (v *ValidatorRemove) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

// Validator leave is used to signal that the sending validator is leaving the network
type ValidatorLeave struct{}

func (v *ValidatorLeave) Type() PayloadType {
	return PayloadTypeValidatorLeave
}

var _ encoding.BinaryUnmarshaler = (*ValidatorLeave)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorLeave)(nil)

func (v *ValidatorLeave) UnmarshalBinary(b []byte) error {
	return serialize.Decode(b, v)
}

func (v *ValidatorLeave) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

// in the future, if/when we go to implement voting based on token weight (instead of validatorship),
// we will create identical payloads as the VoteIDs and VoteBodies payloads, but with different types

// ValidatorVoteIDs is a payload for submitting approvals for any pending resolution, by ID.
type ValidatorVoteIDs struct {
	// ResolutionIDs is an array of all resolution IDs the caller is approving.
	ResolutionIDs []*types.UUID
}

var _ Payload = (*ValidatorVoteIDs)(nil)

func (v *ValidatorVoteIDs) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(v)
}

func (v *ValidatorVoteIDs) Type() PayloadType {
	return PayloadTypeValidatorVoteIDs
}

func (v *ValidatorVoteIDs) UnmarshalBinary(p0 serialize.SerializedData) error {
	return serialize.Decode(p0, v)
}

// ValidatorVoteBodies is a payload for submitting the full vote bodies for any resolution.
type ValidatorVoteBodies struct {
	// Events is an array of the full resolution bodies the caller is voting for.
	Events []*VotableEvent
}

var _ Payload = (*ValidatorVoteBodies)(nil)

// VotableEvent is an event that can be included
// in a ValidatorVoteBodies payload.
type VotableEvent struct {
	Type string
	Body []byte
}

func (v *ValidatorVoteBodies) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(v)
}

func (v *ValidatorVoteBodies) Type() PayloadType {
	return PayloadTypeValidatorVoteBodies
}

func (v *ValidatorVoteBodies) UnmarshalBinary(p0 serialize.SerializedData) error {
	return serialize.Decode(p0, v)
}

// CreateResolution is a payload for creating a new resolution.
type CreateResolution struct {
	Resolution *VotableEvent
}

var _ Payload = (*CreateResolution)(nil)

func (v *CreateResolution) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(v)
}

func (v *CreateResolution) Type() PayloadType {
	return PayloadTypeCreateResolution
}

func (v *CreateResolution) UnmarshalBinary(p0 serialize.SerializedData) error {
	return serialize.Decode(p0, v)
}

// ApproveResolution is a payload for approving on a resolution.
type ApproveResolution struct {
	ResolutionID *types.UUID
}

var _ Payload = (*ApproveResolution)(nil)

func (v *ApproveResolution) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(v)
}

func (v *ApproveResolution) Type() PayloadType {
	return PayloadTypeApproveResolution
}

func (v *ApproveResolution) UnmarshalBinary(p0 serialize.SerializedData) error {
	return serialize.Decode(p0, v)
}

// DeleteResolution is a payload for deleting a resolution.
type DeleteResolution struct {
	ResolutionID *types.UUID
}

var _ Payload = (*DeleteResolution)(nil)

func (d *DeleteResolution) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(d)
}

func (d *DeleteResolution) Type() PayloadType {
	return PayloadTypeDeleteResolution
}

func (d *DeleteResolution) UnmarshalBinary(p0 serialize.SerializedData) error {
	return serialize.Decode(p0, d)
}
