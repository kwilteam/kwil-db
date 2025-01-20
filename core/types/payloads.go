package types

// This file implements the Payload interface for the various payload types.
// Each serialization should have a uint16 version encoded first. Unless
// otherwise noted, the byte ordering used is SerializationByteOrder, which is
// presently little endian.
//
// NOTE: most integers such as lengths are either uint16 or uint32, which is
// somewhat inefficient in most cases. We may consider using varint instead.
// See the encoding/binary package for more information.

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/kwilteam/kwil-db/core/crypto"
)

// PayloadType is the type of payload
type PayloadType string

func (p PayloadType) String() string {
	return string(p)
}

// Payload is the interface that all payloads must implement
// Implementations should use Kwil's serialization package to encode and decode themselves
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	Type() PayloadType
}

const (
	PayloadTypeRawStatement        PayloadType = "raw_statement"
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
	// PayloadTypeDropSchema:          &DropSchema{},
	// PayloadTypeDeploySchema:        &Schema{},
	PayloadTypeRawStatement:        &RawStatement{},
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
	// PayloadTypeDeleteResolution:    &DeleteResolution{},
}

// UnmarshalPayload unmarshals a serialized transaction payload into an instance
// of the type registered for the given PayloadType.
func UnmarshalPayload(payloadType PayloadType, payload []byte) (Payload, error) {
	prototype, have := payloadConcreteTypes[payloadType]
	if !have {
		return nil, errors.New("unknown payload type")
	}

	t := reflect.TypeOf(prototype).Elem() // deref ptr
	elem := reflect.New(t)                // reflect.Type => reflect.Value
	instance := elem.Interface()          // reflect.Type => any

	payloadIface, ok := instance.(Payload)
	if !ok { // should be impossible since payloadConcreteTypes maps to a Payload
		return nil, errors.New("instance not a payload")
	}

	err := payloadIface.UnmarshalBinary(payload)
	if err != nil {
		return nil, err
	}

	return payloadIface, nil
}

// payloadTypes includes native types and types registered from extensions.
var payloadTypes = map[PayloadType]bool{
	PayloadTypeRawStatement:        true,
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

// Valid says if the payload type is known. This does not mean that the node
// will execute the transaction, e.g. not yet activated, or removed.
func (p PayloadType) Valid() bool {
	// native types first for speed
	switch p {
	case PayloadTypeValidatorJoin,
		PayloadTypeValidatorApprove,
		PayloadTypeValidatorRemove,
		PayloadTypeValidatorLeave,
		PayloadTypeTransfer,
		PayloadTypeCreateResolution,
		PayloadTypeApproveResolution,
		PayloadTypeDeleteResolution,
		PayloadTypeRawStatement,
		PayloadTypeExecute,
		// These should not come in user transactions, but they are not invalid
		// payload types in general.
		PayloadTypeValidatorVoteIDs,
		PayloadTypeValidatorVoteBodies:

		return true
	default: // check map that includes registered payloads from extensions
		return payloadTypes[p]
	}
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

// RawStatement is a raw SQL statement that is executed as a transaction
type RawStatement struct {
	Statement  string
	Parameters []*NamedValue
}

type NamedValue struct {
	Name  string
	Value *EncodedValue
}

var _ Payload = (*RawStatement)(nil)

// RawStatement serialization is as follows (using SerializationByteOrder in all
// cases):
//
//   - Two bytes for version (uint16), which is presently 0 (rsVersion).
//   - The statement string is written according to WriteString, which has a
//	   4 byte (uint32) length prefix followed by the bytes of the utf8 string.
//   - The number of parameters is written as a uint16.
//   - For each parameter:
//     - The parameter name is written according to WriteString.
//     - The EncodedValue is serialized according to its MarshalBinary,
//       written according to WriteBytes.

const rsVersion = 0

func (r RawStatement) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(rsVersion)); err != nil {
		return nil, err
	}
	// statement string
	err := WriteString(buf, r.Statement)
	if err != nil {
		return nil, err
	}

	// parameters, max 65535 (uint16)
	numParams := len(r.Parameters)
	if err := binary.Write(buf, SerializationByteOrder, uint16(numParams)); err != nil {
		return nil, err
	}
	for _, param := range r.Parameters {
		// param name string
		err := WriteString(buf, param.Name)
		if err != nil {
			return nil, err
		}
		// EncodedValue
		encValBts, err := param.Value.MarshalBinary()
		if err != nil {
			return nil, err
		}
		err = WriteBytes(buf, encValBts)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (r *RawStatement) UnmarshalBinary(b []byte) error {
	rd := bytes.NewReader(b)

	var version uint16
	if err := binary.Read(rd, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != rsVersion {
		return fmt.Errorf("unsupported version %d", version)
	}

	// statement string
	statement, err := ReadString(rd)
	if err != nil {
		return err
	}

	// parameters
	var numParams uint16
	if err := binary.Read(rd, SerializationByteOrder, &numParams); err != nil {
		return err
	}

	params := make([]*NamedValue, numParams)

	for i := range params {
		// param name string
		name, err := ReadString(rd)
		if err != nil {
			return err
		}

		// EncodedValue
		encValBts, err := ReadBytes(rd)
		if err != nil {
			return err
		}
		var encVal EncodedValue
		if err := encVal.UnmarshalBinary(encValBts); err != nil {
			return err
		}

		params[i] = &NamedValue{
			Name:  name,
			Value: &encVal,
		}
	}

	// only modify the input if no errors
	r.Statement = statement
	r.Parameters = params

	return nil
}

func (r RawStatement) Type() PayloadType {
	return PayloadTypeRawStatement
}

// ActionExecution is the payload that is used to execute an action
type ActionExecution struct {
	DBID      string
	Action    string
	Arguments [][]*EncodedValue
}

var _ Payload = (*ActionExecution)(nil)

func (a ActionExecution) Type() PayloadType {
	return PayloadTypeExecute
}

const aeVersion = 0

// ActionExecution serialization is as follows (using SerializationByteOrder in
// all cases):
//
//   - Two bytes for version (uint16), which is presently 0 (aeVersion).
//   - The DBID string is written according to WriteString, which has a
//	   4 byte length prefix followed by the bytes of the utf8 string.
//   - The Action string is written according to WriteString.
//   - The number of batched calls is written as a uint16.
//   - For each batched call:
//     - The number of arguments is written as a uint16.
//     - Each EncodedValue is serialize according to its MarshalBinary,
//       written according to WriteBytes.

func (a ActionExecution) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(aeVersion)); err != nil {
		return nil, err
	}
	// dbid
	err := WriteString(buf, a.DBID)
	if err != nil {
		return nil, err
	}
	// action string
	err = WriteString(buf, a.Action)
	if err != nil {
		return nil, err
	}

	// arguments
	numCalls := len(a.Arguments)
	if err := binary.Write(buf, SerializationByteOrder, uint16(numCalls)); err != nil {
		return nil, err
	}
	for _, args := range a.Arguments {
		numArgs := len(args)
		if err := binary.Write(buf, SerializationByteOrder, uint16(numArgs)); err != nil {
			return nil, err
		}
		for _, encVal := range args {
			// EncodedValue
			encValBts, err := encVal.MarshalBinary()
			if err != nil {
				return nil, err
			}
			err = WriteBytes(buf, encValBts)
			if err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func (a *ActionExecution) UnmarshalBinary(b []byte) error {
	rd := bytes.NewReader(b)
	var version uint16
	if err := binary.Read(rd, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != aeVersion {
		return fmt.Errorf("unsupported version %d", version)
	}
	// dbid
	dbid, err := ReadString(rd)
	if err != nil {
		return err
	}
	// action string
	action, err := ReadString(rd)
	if err != nil {
		return err
	}

	// arguments
	var numCalls uint16
	if err := binary.Read(rd, SerializationByteOrder, &numCalls); err != nil {
		return err
	}
	args := make([][]*EncodedValue, numCalls)
	for i := range args {
		// arguments
		var numArgs uint16
		if err := binary.Read(rd, SerializationByteOrder, &numArgs); err != nil {
			return err
		}
		args[i] = make([]*EncodedValue, numArgs)
		for j := range args[i] {
			// EncodedValue
			encValBts, err := ReadBytes(rd)
			if err != nil {
				return err
			}
			var ev EncodedValue
			if err := ev.UnmarshalBinary(encValBts); err != nil {
				return err
			}
			args[i][j] = &ev
		}
	}

	a.Action = action
	a.DBID = dbid
	a.Arguments = args

	// ensure all args[i] have same length here or in caller?

	return nil
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

var _ encoding.BinaryUnmarshaler = (*ActionCall)(nil)
var _ encoding.BinaryMarshaler = (*ActionCall)(nil)

const acVersion = 0

// ActionCall serialization is as follows (using SerializationByteOrder in
// all cases):
//
//   - Two bytes for version (uint16), which is presently 0 (acVersion).
//   - The DBID string is written according to WriteString, which has a
//	   4 byte length prefix followed by the bytes of the utf8 string.
//   - The Action string is written according to WriteString.
//   - The number of arguments is written as a uint16.
//   - Each EncodedValue is serialize according to its MarshalBinary, and
//     written according to WriteBytes.

func (ac ActionCall) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(acVersion)); err != nil {
		return nil, err
	}
	// dbid
	err := WriteString(buf, ac.DBID)
	if err != nil {
		return nil, err
	}
	// action string
	err = WriteString(buf, ac.Action)
	if err != nil {
		return nil, err
	}
	// arguments
	numArgs := len(ac.Arguments)
	if err := binary.Write(buf, SerializationByteOrder, uint16(numArgs)); err != nil {
		return nil, err
	}
	for _, arg := range ac.Arguments {
		// EncodedValue
		encValBts, err := arg.MarshalBinary()
		if err != nil {
			return nil, err
		}
		err = WriteBytes(buf, encValBts)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (ac *ActionCall) UnmarshalBinary(b []byte) error {
	rd := bytes.NewReader(b)
	var version uint16
	if err := binary.Read(rd, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != acVersion {
		return fmt.Errorf("unsupported version %d", version)
	}
	// dbid
	dbid, err := ReadString(rd)
	if err != nil {
		return err
	}
	// action string
	action, err := ReadString(rd)
	if err != nil {
		return err
	}

	// arguments
	var numArgs uint16
	if err := binary.Read(rd, SerializationByteOrder, &numArgs); err != nil {
		return err
	}
	args := make([]*EncodedValue, numArgs)
	for i := range args {
		// EncodedValue
		encValBts, err := ReadBytes(rd)
		if err != nil {
			return err
		}
		var ev EncodedValue
		if err := ev.UnmarshalBinary(encValBts); err != nil {
			return err
		}
		args[i] = &ev
	}

	ac.Action = action
	ac.DBID = dbid
	ac.Arguments = args

	return nil
}

// EncodedValue is used to encode a value with its type specified. This is used
// as arguments for actions and procedures.
type EncodedValue struct {
	Type DataType `json:"type"`
	// The double slice handles arrays of encoded values.
	// If there is only one element, the outer slice will have length 1.
	Data [][]byte `json:"data"`
}

const evVersion = 0

// EncodedValue serialization is as follows (using SerializationByteOrder in
// all cases):
//
//   - Two bytes for version (uint16), which is presently 0 (evVersion).
//   - The DataType is serialized according to its MarshalBinary, and the
//     bytes are written according to WriteBytes, which has a 4 byte length
//     prefix followed by the bytes of the data.
//   - The number of elements in the data slice is written as a uint16.
//   - Each element in the data slice is written according to WriteBytes.

func (e EncodedValue) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(evVersion)); err != nil {
		return nil, err
	}
	bts, err := e.Type.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := WriteBytes(buf, bts); err != nil {
		return nil, err
	}
	dataLen := len(e.Data)
	if err := binary.Write(buf, SerializationByteOrder, uint16(dataLen)); err != nil {
		return nil, err
	}
	for _, data := range e.Data {
		err = WriteBytes(buf, data)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (e *EncodedValue) UnmarshalBinary(bts []byte) error {
	buf := bytes.NewBuffer(bts)
	// version uint16
	var version uint16
	if err := binary.Read(buf, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != evVersion {
		return fmt.Errorf("unknown version %d", version)
	}

	typeBytes, err := ReadBytes(buf)
	if err != nil {
		return err
	}
	err = e.Type.UnmarshalBinary(typeBytes)
	if err != nil {
		return err
	}

	var dataLen uint16
	if err := binary.Read(buf, SerializationByteOrder, &dataLen); err != nil {
		return err
	}
	e.Data = make([][]byte, dataLen)
	for i := range dataLen {
		data, err := ReadBytes(buf)
		if err != nil {
			return err
		}
		e.Data[i] = data
	}
	return nil
}

// Decode decodes the encoded value to its native Go type.
func (e *EncodedValue) Decode() (any, error) {
	// decodeScalar decodes a scalar value from a byte slice.
	decodeScalar := func(data []byte, typeName string, isArr bool) (any, error) {
		if data == nil {
			if typeName != NullType.Name {
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
		case TextType.Name:
			return string(data), nil
		case IntType.Name:
			if len(data) != 8 {
				return nil, fmt.Errorf("int must be 8 bytes")
			}
			return int64(binary.BigEndian.Uint64(data)), nil
		case BlobType.Name:
			return data, nil
		case UUIDType.Name:
			if len(data) != 16 {
				return nil, fmt.Errorf("uuid must be 16 bytes")
			}
			var uuid UUID
			copy(uuid[:], data)
			return &uuid, nil
		case BoolType.Name:
			return data[0] == 1, nil
		case NullType.Name:
			return nil, nil
		case Uint256Type.Name:
			return Uint256FromBytes(data)
		case NumericStr:
			return ParseDecimal(string(data))
		default:
			return nil, fmt.Errorf("cannot decode type %s", typeName)
		}
	}

	if e.Type.IsArray {
		var arrAny any

		// postgres requires arrays to be of the correct type, not of []any
		switch e.Type.Name {
		case NullType.Name:
			return nil, fmt.Errorf("cannot decode array of type 'null'")
		case TextType.Name:
			arr := make([]string, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(string))
			}
			arrAny = arr
		case IntType.Name:
			arr := make([]int64, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(int64))
			}
			arrAny = arr
		case BlobType.Name:
			arr := make([][]byte, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.([]byte))
			}
			arrAny = arr
		case UUIDType.Name:
			arr := make(UUIDArray, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(*UUID))
			}
			arrAny = arr
		case BoolType.Name:
			arr := make([]bool, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(bool))
			}
			arrAny = arr
		case Uint256Type.Name:
			arr := make(Uint256Array, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(*Uint256))
			}
			arrAny = arr
		case NumericStr:
			arr := make(DecimalArray, 0, len(e.Data))
			for _, elem := range e.Data {
				dec, err := decodeScalar(elem, e.Type.Name, true)
				if err != nil {
					return nil, err
				}

				arr = append(arr, dec.(*Decimal))
			}
			arrAny = arr
		default:
			return nil, fmt.Errorf("unknown type `%s`", e.Type.Name)
		}
		return arrAny, nil
	}

	if e.Type.Name == NullType.Name {
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
				Name: NullType.Name,
			},
			Data: nil,
		}, nil
	}

	// encodeScalar encodes a scalar value into a byte slice.
	// It also returns the data type of the value.
	encodeScalar := func(v any) ([]byte, *DataType, error) {
		switch t := v.(type) {
		case string:
			return []byte(t), TextType, nil
		case int, int16, int32, int64, int8, uint, uint16, uint32, uint64: // intentionally ignore uint8 since it is an alias for byte
			i64, err := strconv.ParseInt(fmt.Sprint(t), 10, 64)
			if err != nil {
				return nil, nil, err
			}

			var buf [8]byte
			binary.BigEndian.PutUint64(buf[:], uint64(i64))
			return buf[:], IntType, nil
		case []byte:
			return t, BlobType, nil
		case [16]byte:
			return t[:], UUIDType, nil
		case UUID:
			return t[:], UUIDType, nil
		case *UUID:
			return t[:], UUIDType, nil
		case bool:
			if t {
				return []byte{1}, BoolType, nil
			}
			return []byte{0}, BoolType, nil
		case nil: // since we quick return for nil, we can only reach this point if the type is nil
			// and we are in an array
			return nil, nil, fmt.Errorf("cannot encode nil in type array")
		case *Decimal:
			decTyp, err := NewNumericType(t.Precision(), t.Scale())
			if err != nil {
				return nil, nil, err
			}

			return []byte(t.String()), decTyp, nil
		case Decimal:
			decTyp, err := NewNumericType(t.Precision(), t.Scale())
			if err != nil {
				return nil, nil, err
			}

			return []byte(t.String()), decTyp, nil
		case *Uint256:
			return t.Bytes(), Uint256Type, nil
		case Uint256:
			return t.Bytes(), Uint256Type, nil
		default:
			return nil, nil, fmt.Errorf("cannot encode type %T", v)
		}
	}

	dt := &DataType{}
	// check if it is an array
	typeOf := reflect.TypeOf(v)
	if typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() != reflect.Uint8 { // ignore byte slices
		// encode each element of the array
		encoded := make([][]byte, 0)
		// it can be of any slice type, e.g. []any, []string, []int, etc.
		valueOf := reflect.ValueOf(v)
		for i := range valueOf.Len() {
			elem := valueOf.Index(i).Interface()
			enc, t, err := encodeScalar(elem)
			if err != nil {
				return nil, err
			}

			if !t.EqualsStrict(NullType) {
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
			dt.Name = NullType.Name
		}

		dt.IsArray = true

		return &EncodedValue{
			Type: *dt,
			Data: encoded,
		}, nil
	}

	enc, t, err := encodeScalar(v)
	if err != nil {
		return nil, err
	}

	return &EncodedValue{
		Type: *t,
		Data: [][]byte{enc},
	}, nil
}

// Transfer transfers an amount of tokens from the sender to the receiver.
type Transfer struct {
	To     *AccountID `json:"to"`     // to be string as user identifier
	Amount *big.Int   `json:"amount"` // big.Int
}

var _ Payload = (*Transfer)(nil)

func (v Transfer) Type() PayloadType {
	return PayloadTypeTransfer
}

var _ encoding.BinaryUnmarshaler = (*Transfer)(nil)
var _ encoding.BinaryMarshaler = (*Transfer)(nil)
var _ encoding.BinaryMarshaler = Transfer{}

// transfer payload version
const tVersion = 0

func (v Transfer) MarshalBinary() ([]byte, error) {
	if v.To == nil {
		return nil, errors.New("missing To field in transfer")
	}

	buf := new(bytes.Buffer)

	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(tVersion)); err != nil {
		return nil, err
	}

	// transfer to
	toBts, err := v.To.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if err := WriteBytes(buf, toBts); err != nil {
		return nil, err
	}

	// transfer amount
	if err := WriteBigInt(buf, v.Amount); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *Transfer) UnmarshalBinary(b []byte) error {
	rd := bytes.NewReader(b)

	var version uint16
	if err := binary.Read(rd, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != tVersion {
		return fmt.Errorf("unsupported transfer payload version %d", version)
	}

	// transfer to
	toBts, err := ReadBytes(rd)
	if err != nil {
		return err
	}

	v.To = &AccountID{}
	err = v.To.UnmarshalBinary(toBts)
	if err != nil {
		return err
	}

	// transfer amount
	v.Amount, err = ReadBigInt(rd)
	if err != nil {
		return err
	}
	return nil
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

const vjVersion = 0

func (v ValidatorJoin) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2+8)
	SerializationByteOrder.PutUint16(b, vjVersion)
	SerializationByteOrder.PutUint64(b[2:], v.Power)
	return b, nil
}

func (v *ValidatorJoin) UnmarshalBinary(b []byte) error {
	if len(b) < 2+8 {
		return fmt.Errorf("invalid length %d", len(b))
	}

	version := SerializationByteOrder.Uint16(b)
	if version != vjVersion {
		return fmt.Errorf("invalid version %d", version)
	}
	v.Power = SerializationByteOrder.Uint64(b[2:])
	return nil
}

// ValidatorApprove is used to vote for a validators approval to join the network
type ValidatorApprove struct {
	Candidate []byte
	KeyType   crypto.KeyType
}

func (v *ValidatorApprove) Type() PayloadType {
	return PayloadTypeValidatorApprove
}

var _ encoding.BinaryUnmarshaler = (*ValidatorApprove)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorApprove)(nil)

// UnmarshalBinary and MarshalBinary in the same manner as ValidatorRemove

const vaVersion = 0

func (v ValidatorApprove) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, SerializationByteOrder, uint16(vaVersion))
	WriteBytes(buf, v.Candidate)
	binary.Write(buf, SerializationByteOrder, int32(v.KeyType))
	return buf.Bytes(), nil
}

func (v *ValidatorApprove) UnmarshalBinary(b []byte) error {
	rd := bytes.NewReader(b)
	var version uint16
	err := binary.Read(rd, SerializationByteOrder, &version)
	if err != nil {
		return err
	}
	if version != vrVersion {
		return fmt.Errorf("invalid validator remove payload version")
	}
	candidate, err := ReadBytes(rd)
	if err != nil {
		return err
	}
	var keyType int32
	err = binary.Read(rd, SerializationByteOrder, &keyType)
	if err != nil {
		return err
	}
	kt := crypto.KeyType(keyType)
	if !kt.Valid() {
		return fmt.Errorf("invalid key type")
	}

	v.Candidate = candidate
	v.KeyType = kt

	return nil
}

// ValidatorRemove is used to vote for a validators removal from the network
type ValidatorRemove struct {
	Validator []byte
	KeyType   crypto.KeyType
}

func (v *ValidatorRemove) Type() PayloadType {
	return PayloadTypeValidatorRemove
}

var _ encoding.BinaryUnmarshaler = (*ValidatorRemove)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorRemove)(nil)

const vrVersion = 0

func (v ValidatorRemove) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, SerializationByteOrder, uint16(vrVersion))

	WriteBytes(buf, v.Validator)

	binary.Write(buf, SerializationByteOrder, int32(v.KeyType))

	return buf.Bytes(), nil
}

func (v *ValidatorRemove) UnmarshalBinary(b []byte) error {
	rd := bytes.NewReader(b)
	var version uint16
	err := binary.Read(rd, SerializationByteOrder, &version)
	if err != nil {
		return err
	}
	if version != vrVersion {
		return fmt.Errorf("invalid validator remove payload version")
	}
	val, err := ReadBytes(rd)
	if err != nil {
		return err
	}

	var keyType int32
	err = binary.Read(rd, SerializationByteOrder, &keyType)
	if err != nil {
		return err
	}
	kt := crypto.KeyType(keyType)
	if !kt.Valid() {
		return fmt.Errorf("invalid key type")
	}

	v.Validator = val
	v.KeyType = kt

	return nil
}

// Validator leave is used to signal that the sending validator is leaving the network
type ValidatorLeave struct{}

func (v *ValidatorLeave) Type() PayloadType {
	return PayloadTypeValidatorLeave
}

var _ encoding.BinaryUnmarshaler = (*ValidatorLeave)(nil)
var _ encoding.BinaryMarshaler = (*ValidatorLeave)(nil)
var _ encoding.BinaryMarshaler = ValidatorLeave{}

const vlVersion = 0

func (v ValidatorLeave) MarshalBinary() ([]byte, error) {
	// just a version uint16 and that's all
	return SerializationByteOrder.AppendUint16(nil, vlVersion), nil
}

func (v *ValidatorLeave) UnmarshalBinary(b []byte) error {
	if len(b) != 2 {
		return fmt.Errorf("invalid validator leave payload")
	}
	if SerializationByteOrder.Uint16(b) != vlVersion {
		return fmt.Errorf("invalid validator leave payload version")
	}
	return nil
}

// in the future, if/when we go to implement voting based on token weight (instead of validatorship),
// we will create identical payloads as the VoteIDs and VoteBodies payloads, but with different types

// ValidatorVoteIDs is a payload for submitting approvals for any pending resolution, by ID.
type ValidatorVoteIDs struct {
	// ResolutionIDs is an array of all resolution IDs the caller is approving.
	ResolutionIDs []*UUID
}

var _ Payload = (*ValidatorVoteIDs)(nil)

const vvidVersion = 0

func (v *ValidatorVoteIDs) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(vvidVersion)); err != nil {
		return nil, err
	}

	// Length of resolution IDs (uint32)
	if err := binary.Write(buf, SerializationByteOrder, uint32(len(v.ResolutionIDs))); err != nil {
		return nil, err
	}

	for _, id := range v.ResolutionIDs {
		enc, err := id.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if err := WriteBytes(buf, enc); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (v *ValidatorVoteIDs) UnmarshalBinary(bts []byte) error {
	buf := bytes.NewBuffer(bts)
	var version uint16
	if err := binary.Read(buf, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != vvidVersion {
		return fmt.Errorf("unknown version: %d", version)
	}
	var length uint32
	if err := binary.Read(buf, SerializationByteOrder, &length); err != nil {
		return err
	}
	v.ResolutionIDs = make([]*UUID, 0, length) // to match MArshalBinary
	for range length {
		idBts, err := ReadBytes(buf)
		if err != nil {
			return err
		}
		id := &UUID{}
		if err := id.UnmarshalBinary(idBts); err != nil {
			return err
		}
		v.ResolutionIDs = append(v.ResolutionIDs, id)
	}
	return nil
}

func (v *ValidatorVoteIDs) Type() PayloadType {
	return PayloadTypeValidatorVoteIDs
}

// ValidatorVoteBodies is a payload for submitting the full vote bodies for any resolution.
type ValidatorVoteBodies struct {
	// Events is an array of the full resolution bodies the caller is voting for.
	Events []*VotableEvent
}

var _ Payload = (*ValidatorVoteBodies)(nil)

func (v *ValidatorVoteBodies) Type() PayloadType {
	return PayloadTypeValidatorVoteBodies
}

const vvbbVersion = 0

func (v ValidatorVoteBodies) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	// version uint16
	if err := binary.Write(buf, SerializationByteOrder, uint16(vvbbVersion)); err != nil {
		return nil, err
	}

	// Length of events (uint32)
	if err := binary.Write(buf, SerializationByteOrder, uint32(len(v.Events))); err != nil {
		return nil, err
	}
	for _, event := range v.Events {
		evtBts, err := event.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if err := WriteBytes(buf, evtBts); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (v *ValidatorVoteBodies) UnmarshalBinary(bts []byte) error {
	buf := bytes.NewBuffer(bts)
	var version uint16
	if err := binary.Read(buf, SerializationByteOrder, &version); err != nil {
		return err
	}
	if version != vvbbVersion {
		return fmt.Errorf("unknown version: %d", version)
	}
	var numEvents uint32
	if err := binary.Read(buf, SerializationByteOrder, &numEvents); err != nil {
		return err
	}
	if int(numEvents) > min(500_000, buf.Len()) {
		return fmt.Errorf("invalid event count: %d", numEvents)
	}
	v.Events = make([]*VotableEvent, numEvents)
	for i := range v.Events {
		evtBts, err := ReadBytes(buf)
		if err != nil {
			return err
		}
		event := &VotableEvent{}
		if err := event.UnmarshalBinary(evtBts); err != nil {
			return err
		}
		v.Events[i] = event
	}
	return nil
}

// CreateResolution is a payload for creating a new resolution.
type CreateResolution struct {
	Resolution *VotableEvent
}

var _ Payload = (*CreateResolution)(nil)

const crVersion = 0

func (v CreateResolution) MarshalBinary() ([]byte, error) {
	// version uint16 and then the v.Resolution.MarshalBinary
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, SerializationByteOrder, uint16(crVersion)); err != nil {
		return nil, err
	}
	enc, err := v.Resolution.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := binary.Write(buf, SerializationByteOrder, enc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *CreateResolution) UnmarshalBinary(bts []byte) error {
	if len(bts) <= 2 {
		return fmt.Errorf("invalid payload")
	}
	version := SerializationByteOrder.Uint16(bts)
	if version != crVersion {
		return fmt.Errorf("unknown version: %d", version)
	}
	// use buf[2:] to unmarshal the Resolution
	var resolution VotableEvent
	if err := resolution.UnmarshalBinary(bts[2:]); err != nil {
		return err
	}
	v.Resolution = &resolution
	return nil
}

func (v *CreateResolution) Type() PayloadType {
	return PayloadTypeCreateResolution
}

// ApproveResolution is a payload for approving on a resolution.
type ApproveResolution struct {
	ResolutionID *UUID
}

var _ Payload = (*ApproveResolution)(nil)

const arVersion = 0

func (v ApproveResolution) MarshalBinary() ([]byte, error) {
	// uint16 version and then the v.ResolutionID.MarshalBinary
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, SerializationByteOrder, uint16(arVersion)); err != nil {
		return nil, err
	}
	// var resID UUID
	// if v.ResolutionID != nil {
	// 	resID = *v.ResolutionID
	// }
	if v.ResolutionID == nil {
		return nil, fmt.Errorf("resolution ID is nil")
	}
	enc, err := v.ResolutionID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := binary.Write(buf, SerializationByteOrder, enc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *ApproveResolution) UnmarshalBinary(bts []byte) error {
	if len(bts) <= 2 {
		return fmt.Errorf("invalid payload")
	}
	version := SerializationByteOrder.Uint16(bts)
	if version != arVersion {
		return fmt.Errorf("unknown version: %d", version)
	}
	// use buf[2:] to unmarshal the ResolutionID
	var resolutionID UUID
	if err := resolutionID.UnmarshalBinary(bts[2:]); err != nil {
		return err
	}
	v.ResolutionID = &resolutionID
	return nil
}

func (v *ApproveResolution) Type() PayloadType {
	return PayloadTypeApproveResolution
}

// DeleteResolution is a payload for deleting a resolution.
type DeleteResolution struct {
	ResolutionID *UUID
}

var _ Payload = (*DeleteResolution)(nil)

const drVersion = 0

func (d DeleteResolution) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, SerializationByteOrder, uint16(drVersion)); err != nil {
		return nil, err
	}
	if d.ResolutionID == nil {
		return nil, fmt.Errorf("resolution ID is nil")
	}
	enc, err := d.ResolutionID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := binary.Write(buf, SerializationByteOrder, enc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *DeleteResolution) Type() PayloadType {
	return PayloadTypeDeleteResolution
}

func (d *DeleteResolution) UnmarshalBinary(bts []byte) error {
	if len(bts) <= 2 {
		return fmt.Errorf("invalid payload")
	}
	version := SerializationByteOrder.Uint16(bts)
	if version != drVersion {
		return fmt.Errorf("unknown version: %d", version)
	}
	// use buf[2:] to unmarshal the ResolutionID
	var resolutionID UUID
	if err := resolutionID.UnmarshalBinary(bts[2:]); err != nil {
		return err
	}
	d.ResolutionID = &resolutionID
	return nil
}
