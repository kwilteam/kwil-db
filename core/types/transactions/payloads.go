package transactions

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
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
	Arguments [][]*EncodedValue
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

// ActionCall models the arguments of an action call. It would be serialized
// into CallMessage.Body. This is not a transaction payload. See
// transactions.ActionExecution for the transaction payload used for executing
// an action.
type ActionCall struct {
	DBID      string
	Action    string
	Arguments []*EncodedValue
}

func (a *ActionCall) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(a)
}

func (a *ActionCall) UnmarshalBinary(b serialize.SerializedData) error {
	return serialize.Decode(b, a)
}

var _ encoding.BinaryUnmarshaler = (*ActionCall)(nil)
var _ encoding.BinaryMarshaler = (*ActionCall)(nil)

// EncodedValue is used to encode a value with its type specified
type EncodedValue struct {
	Type DataType
	// The double slice handles arrays of encoded values.
	// If there is only one element, the outer slice will have length 1.
	Data [][]byte `rlp:"optional"`
}

// Decode decodes the encoded value to its native Go type.
func (e *EncodedValue) Decode() (any, error) {
	// decodeScalar decodes a scalar value from a byte slice.
	decodeScalar := func(data []byte, typeName string, isArr bool) (any, error) {
		if data == nil {
			if typeName != types.NullType.Name {
				// this is not super clean, but gives a much more helpful error message
				pref := ""
				if isArr {
					pref = "[]"
				}
				return nil, fmt.Errorf("cannot decode nil data into type %s"+pref, typeName)
			}
			return nil, nil
		}

		switch typeName {
		case types.TextType.Name:
			return string(data), nil
		case types.IntType.Name:
			if len(data) != 8 {
				return nil, fmt.Errorf("int must be 8 bytes")
			}
			return int64(binary.BigEndian.Uint64(data)), nil
		case types.BlobType.Name:
			return data, nil
		case types.UUIDType.Name:
			if len(data) != 16 {
				return nil, fmt.Errorf("uuid must be 16 bytes")
			}
			var uuid types.UUID
			copy(uuid[:], data)
			return &uuid, nil
		case types.BoolType.Name:
			return data[0] == 1, nil
		case types.NullType.Name:
			return nil, nil
		case types.Uint256Type.Name:
			return types.Uint256FromBytes(data)
		case types.DecimalStr:
			return decimal.NewFromString(string(data))
		default:
			return nil, fmt.Errorf("cannot decode type %s", typeName)
		}
	}

	if e.Type.IsArray {
		var arrAny any

		// postgres requires arrays to be of the correct type, not of []any
		switch e.Type.Name {
		case types.NullType.Name:
			return nil, fmt.Errorf("cannot decode array of type 'null'")
		case types.TextType.Name:
			arr := make([]string, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(string))
			}
			arrAny = arr
		case types.IntType.Name:
			arr := make([]int64, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(int64))
			}
			arrAny = arr
		case types.BlobType.Name:
			arr := make([][]byte, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.([]byte))
			}
			arrAny = arr
		case types.UUIDType.Name:
			arr := make(types.UUIDArray, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(*types.UUID))
			}
			arrAny = arr
		case types.BoolType.Name:
			arr := make([]bool, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(bool))
			}
			arrAny = arr
		case types.Uint256Type.Name:
			arr := make(types.Uint256Array, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(*types.Uint256))
			}
			arrAny = arr
		case types.DecimalStr:
			arr := make(decimal.DecimalArray, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(*decimal.Decimal))
			}
			arrAny = arr
		default:
			return nil, fmt.Errorf("unknown type `%s`", e.Type.Name)
		}
		return arrAny, nil
	}

	if e.Type.Name == types.NullType.Name {
		return nil, nil
	}
	if len(e.Data) != 1 {
		return nil, fmt.Errorf("expected 1 element, got %d", len(e.Data))
	}

	return decodeScalar(e.Data[0], e.Type.Name, false)
}

// EncodeValue encodes a value to its detected type.
// It will reflect the value of the passed argument to determine its type.
func EncodeValue(v any) (*EncodedValue, error) {
	if v == nil {
		return &EncodedValue{
			Type: DataType{
				Name: types.NullType.Name,
			},
			Data: nil,
		}, nil
	}

	// encodeScalar encodes a scalar value into a byte slice.
	// It also returns the data type of the value.
	encodeScalar := func(v any) ([]byte, *types.DataType, error) {
		switch t := v.(type) {
		case string:
			return []byte(t), types.TextType, nil
		case int, int16, int32, int64, int8, uint, uint16, uint32, uint64: // intentionally ignore uint8 since it is an alias for byte
			i64, err := strconv.ParseInt(fmt.Sprint(t), 10, 64)
			if err != nil {
				return nil, nil, err
			}

			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(i64))
			return buf[:], types.IntType, nil
		case []byte:
			return t, types.BlobType, nil
		case [16]byte:
			return t[:], types.UUIDType, nil
		case types.UUID:
			return t[:], types.UUIDType, nil
		case *types.UUID:
			return t[:], types.UUIDType, nil
		case bool:
			if t {
				return []byte{1}, types.BoolType, nil
			}
			return []byte{0}, types.BoolType, nil
		case nil: // since we quick return for nil, we can only reach this point if the type is nil
			// and we are in an array
			return nil, nil, fmt.Errorf("cannot encode nil in type array")
		case *decimal.Decimal:
			decTyp, err := types.NewDecimalType(t.Precision(), t.Scale())
			if err != nil {
				return nil, nil, err
			}

			return []byte(t.String()), decTyp, nil
		case decimal.Decimal:
			decTyp, err := types.NewDecimalType(t.Precision(), t.Scale())
			if err != nil {
				return nil, nil, err
			}

			return []byte(t.String()), decTyp, nil
		case *types.Uint256:
			return t.Bytes(), types.Uint256Type, nil
		case types.Uint256:
			return t.Bytes(), types.Uint256Type, nil
		default:
			return nil, nil, fmt.Errorf("cannot encode type %T", v)
		}
	}

	dt := &types.DataType{}
	// check if it is an array
	typeOf := reflect.TypeOf(v)
	if typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() != reflect.Uint8 { // ignore byte slices
		// encode each element of the array
		encoded := make([][]byte, 0)
		// it can be of any slice type, e.g. []any, []string, []int, etc.
		valueOf := reflect.ValueOf(v)
		for i := 0; i < valueOf.Len(); i++ {
			elem := valueOf.Index(i).Interface()
			enc, t, err := encodeScalar(elem)
			if err != nil {
				return nil, err
			}

			if !t.EqualsStrict(types.NullType) {
				if dt.Name == "" {
					*dt = *t
				} else if !dt.EqualsStrict(t) {
					return nil, fmt.Errorf("array contains elements of different types")
				}
			}

			encoded = append(encoded, enc)
		}

		// edge case where all elements are nil
		if dt.Name == "" {
			dt.Name = types.NullType.Name
		}

		dt.IsArray = true

		localDt := DataType{}
		localDt.fromTypes(dt)
		return &EncodedValue{
			Type: localDt,
			Data: encoded,
		}, nil
	}

	enc, t, err := encodeScalar(v)
	if err != nil {
		return nil, err
	}

	localDt := DataType{}
	localDt.fromTypes(t)
	return &EncodedValue{
		Type: localDt,
		Data: [][]byte{enc},
	}, nil
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
