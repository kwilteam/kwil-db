package types

// This file defines the EncodeValue type used to represent the arguments to
// action/procedure calls (read-only) or execution (from transactions).

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"

	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// EncodedValue is used to encode a value with its type specified. These are
// used as the arguments for action calls and action execution. Create an
// EncodedValue with EncodeValue.
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
		case DecimalStr:
			return decimal.NewFromString(string(data))
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
		case DecimalStr:
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
		case *decimal.Decimal:
			decTyp, err := NewDecimalType(t.Precision(), t.Scale())
			if err != nil {
				return nil, nil, err
			}

			return []byte(t.String()), decTyp, nil
		case decimal.Decimal:
			decTyp, err := NewDecimalType(t.Precision(), t.Scale())
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
		for i := 0; i < valueOf.Len(); i++ {
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
