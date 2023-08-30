package transactions

import (
	"encoding"

	"github.com/kwilteam/kwil-db/pkg/serialize"
)

// PayloadType is the type of payload
type PayloadType string

func (p PayloadType) String() string {
	return string(p)
}

func (p PayloadType) Valid() bool {
	switch p {
	case PayloadTypeDeploySchema,
		PayloadTypeDropSchema,
		PayloadTypeExecuteAction,
		PayloadTypeCallAction:
		return true
	default:
		return false
	}
}

const (
	PayloadTypeDeploySchema     PayloadType = "deploy_schema"
	PayloadTypeDropSchema       PayloadType = "drop_schema"
	PayloadTypeExecuteAction    PayloadType = "execute_action"
	PayloadTypeCallAction       PayloadType = "call_action"
	PayloadTypeValidatorJoin    PayloadType = "validator_join"
	PayloadTypeValidatorLeave   PayloadType = "validator_leave"
	PayloadTypeValidatorApprove PayloadType = "validator_approve"
)

// Payload is the interface that all payloads must implement
// Implementations should use Kwil's serialization package to encode and decode themselves
type Payload interface {
	MarshalBinary() (serialize.SerializedData, error)
	UnmarshalBinary(serialize.SerializedData) error
	Type() PayloadType
}

// Schema is the payload that is used to deploy a schema
type Schema struct {
	Owner      []byte       // public key of the owner
	Name       string       `json:"name"`
	Tables     []*Table     `json:"tables"`
	Actions    []*Action    `json:"actions"`
	Extensions []*Extension `json:"extensions"`
}

var _ Payload = (*Schema)(nil)

func (s *Schema) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(s)
}

func (s *Schema) UnmarshalBinary(b serialize.SerializedData) error {
	result, err := serialize.Decode[Schema](b)
	if err != nil {
		return err
	}

	*s = *result
	return nil
}

func (s *Schema) Type() PayloadType {
	return PayloadTypeDeploySchema
}

type Extension struct {
	Name   string             `json:"name"`
	Config []*ExtensionConfig `json:"config"`
	Alias  string             `json:"alias"`
}

type ExtensionConfig struct {
	Argument string `json:"argument"`
	Value    string `json:"value"`
}

type Table struct {
	Name        string        `json:"name"`
	Columns     []*Column     `json:"columns,omitempty"`
	Indexes     []*Index      `json:"indexes,omitempty"`
	ForeignKeys []*ForeignKey `json:"foreign_keys,omitempty"`
}

type Column struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Attributes []*Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

type Action struct {
	Name   string   `json:"name"`
	Inputs []string `json:"inputs,omitempty"`
	// Mutability could be empty if the abi is generated by legacy version of kuneiform,
	// default to "update" for backward compatibility
	Mutability string `json:"mutability,omitempty"`
	// Auxiliaries are the auxiliary types that are required for the action, specifying extra characteristics of the action
	Auxiliaries []string `json:"auxiliaries,omitempty"`
	Public      bool     `json:"public,omitempty"`
	Statements  []string `json:"statements,omitempty"`
}

type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
}

type ForeignKey struct {
	// ChildKeys are the columns that are referencing another.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "a" is the child key
	ChildKeys []string `json:"child_keys"`

	// ParentKeys are the columns that are being referred to.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2.b" is the parent key
	ParentKeys []string `json:"parent_keys"`

	// ParentTable is the table that holds the parent columns.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2.b" is the parent table
	ParentTable string `json:"parent_table"`

	// Action refers to what the foreign key should do when the parent is altered.
	// This is NOT the same as a database action;
	// however sqlite's docs refer to these as actions,
	// so we should be consistent with that.
	// For example, ON DELETE CASCADE is a foreign key action
	Actions []*ForeignKeyAction `json:"actions"`
}

// ForeignKeyAction is used to specify what should occur
// if a parent key is updated or deleted
type ForeignKeyAction struct {
	// On can be either "UPDATE" or "DELETE"
	On string `json:"on"`

	// Do specifies what a foreign key action should do
	Do string `json:"do"`
}

// MutabilityType is the type of mutability
type MutabilityType string

func (t MutabilityType) String() string {
	return string(t)
}

const (
	MutabilityUpdate MutabilityType = "update"
	MutabilityView   MutabilityType = "view"
)

// AuxiliaryType is the type of auxiliary
type AuxiliaryType string

func (t AuxiliaryType) String() string {
	return string(t)
}

const (
	// AuxiliaryTypeMustSign is used to specify that an action need signature, it is used for `view` action.
	AuxiliaryTypeMustSign AuxiliaryType = "mustsign"
	// AuxiliaryTypeOwner is used to specify that an action caller must be the owner of the database.
	AuxiliaryTypeOwner AuxiliaryType = "owner"
)

// DropSchema is the payload that is used to drop a schema
type DropSchema struct {
	DBID string
}

var _ Payload = (*DropSchema)(nil)

func (s *DropSchema) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(s)
}

func (s *DropSchema) UnmarshalBinary(b serialize.SerializedData) error {
	res, err := serialize.Decode[DropSchema](b)
	if err != nil {
		return err
	}

	*s = *res

	return nil
}

func (s *DropSchema) Type() PayloadType {
	return PayloadTypeDropSchema
}

// ActionExecution is the payload that is used to execute an action
type ActionExecution struct {
	DBID      string
	Action    string
	Arguments [][]string
}

var _ Payload = (*ActionExecution)(nil)

func (a *ActionExecution) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(a)
}

func (a *ActionExecution) UnmarshalBinary(b serialize.SerializedData) error {
	res, err := serialize.Decode[ActionExecution](b)
	if err != nil {
		return err
	}

	*a = *res
	return nil
}

func (a *ActionExecution) Type() PayloadType {
	return PayloadTypeExecuteAction
}

// ActionCall is the payload that is used to call an action
type ActionCall struct {
	DBID      string
	Action    string
	Arguments []string
}

var _ Payload = (*ActionCall)(nil)

func (a *ActionCall) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(a)
}

func (a *ActionCall) UnmarshalBinary(b serialize.SerializedData) error {
	res, err := serialize.Decode[ActionCall](b)
	if err != nil {
		return err
	}

	*a = *res
	return nil
}

var _ encoding.BinaryUnmarshaler = (*ActionCall)(nil)
var _ encoding.BinaryMarshaler = (*ActionCall)(nil)

func (a *ActionCall) Type() PayloadType {
	return PayloadTypeCallAction
}

type ValidatorJoin struct {
	Candidate []byte
	Power     uint64
}

func (v *ValidatorJoin) Type() PayloadType {
	return PayloadTypeValidatorJoin
}

var _ encoding.BinaryUnmarshaler = (*ValidatorJoin)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorJoin)(nil)

func (v *ValidatorJoin) UnmarshalBinary(b []byte) error {
	return serialize.DecodeInto(b, v)
}

func (v *ValidatorJoin) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

type ValidatorApprove struct {
	Candidate []byte
}

func (v *ValidatorApprove) Type() PayloadType {
	return PayloadTypeValidatorApprove
}

var _ encoding.BinaryUnmarshaler = (*ValidatorApprove)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorApprove)(nil)

func (v *ValidatorApprove) UnmarshalBinary(b []byte) error {
	return serialize.DecodeInto(b, v)
}

func (v *ValidatorApprove) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}

type ValidatorLeave struct {
	Validator []byte
}

func (v *ValidatorLeave) Type() PayloadType {
	return PayloadTypeValidatorLeave
}

var _ encoding.BinaryUnmarshaler = (*ValidatorLeave)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorLeave)(nil)

func (v *ValidatorLeave) UnmarshalBinary(b []byte) error {
	return serialize.DecodeInto(b, v)
}

func (v *ValidatorLeave) MarshalBinary() ([]byte, error) {
	return serialize.Encode(v)
}
