package pg

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

func init() {
	registerDatatype(textType, textArrayType)
	registerDatatype(intType, intArrayType)
	registerDatatype(boolType, boolArrayType)
	registerDatatype(blobType, blobArrayType)
	registerDatatype(uuidType, uuidArrayType)
	registerDatatype(decimalType, decimalArrayType)
	registerDatatype(uint256Type, uint256ArrayType)
}

var (
	dataTypesByMatch   = map[reflect.Type]*datatype{}
	scalarToArray      = map[*datatype]*datatype{} // maps the scalar type to the array type
	datatypes          = map[*datatype]struct{}{}  // a set of all data types (used for iteration)
	kwilTypeToDataType = map[types.DataType]*datatype{}
)

// registerOIDs registers all of the data types that we support in Postgres.
func registerDatatype(scalar *datatype, array *datatype) {
	for _, match := range scalar.Matches {
		_, ok := dataTypesByMatch[match]
		if ok {
			panic(fmt.Sprintf("data type %T already registered", match))
		}

		dataTypesByMatch[match] = scalar
		datatypes[scalar] = struct{}{}
	}

	for _, match := range array.Matches {
		_, ok := dataTypesByMatch[match]
		if ok {
			panic(fmt.Sprintf("data type %T already registered", match))
		}

		dataTypesByMatch[match] = array
		datatypes[array] = struct{}{}
	}

	_, ok := kwilTypeToDataType[*scalar.KwilType]
	if ok {
		k := kwilTypeToDataType
		_ = k
		panic(fmt.Sprintf("Kwil type %s already registered", scalar.KwilType.String()))
	}

	kwilTypeToDataType[*scalar.KwilType] = scalar

	_, ok = kwilTypeToDataType[*array.KwilType]
	if ok {
		panic(fmt.Sprintf("Kwil type %s already registered", array.KwilType.String()))
	}

	kwilTypeToDataType[*array.KwilType] = array

	scalarToArray[scalar] = array
}

// datatype allows us to easily register new data types.
// It is used to define how to encode and decode data types in Postgres.
// While all of the implementations for this are stored in the PG package,
// the primary reason for identifying this as an interface is to allow for
// easy addition of types in the future (knowing what needs to be implemented
// to support new data types).
type datatype struct {
	// KwilType is the Kwil-native data type that is tied to this data type.
	// There must be exaclty one. It will ignore all metadata (e.g. for decimal, any
	// precision/scale is ignore).
	KwilType *types.DataType
	// Matches is the list of all data types that this type matches.
	// These will be stored in a map, and thus each match type can only be
	// used once across all data types.
	Matches []reflect.Type
	// OID returns the OID of the data type in Postgres.
	// It will be given to Postgres when encoding the data type
	// with QueryModeInferredArgTypes, and will also be used to identify
	// how values should be decoded.
	OID func(*pgtype.Map) uint32
	// ExtraOIDs returns any additional OIDs which the data type can be decoded from.
	// This is useful for int types, which can be decoded from int2, int4, and int8.
	// These will be used in addition to the OID returned by OID.
	// This can be nil if there are no additional OIDs.
	ExtraOIDs []uint32
	// EncodeInferred encodes a value into a byte slice, given the type of the value.
	// The passed value will always be of a type that matches one of the Matches types.
	// It must return the serialized data.
	// This is used when operating in QueryModeInferredArgTypes, to infer the postgres
	// data type from the native go type.
	// If not using QueryModeInferredArgTypes, it will be encoded using a driver.Valuer,
	// or as a native go type.
	EncodeInferred func(any) (any, error)
	// Decode decodes a data type received from Postgres. The input will either be a data type
	// native to Go, a type defined in pgx, or a type in a custom pgx Codec (which we currently
	// don't use).
	Decode func(any) (any, error)
	// SerializeChangeset decodes a data type received from Postgres as a string. PGX only returns
	// replication data as strings, so this is used to decode replication data. Decode will never be called
	// with null values, but it may be called with empty strings / 0 values.
	// https://github.com/jackc/pglogrepl/blob/828fbfe908e97cfeb409a17e4ec339dede1f1a17/message.go#L379
	SerializeChangeset func(value string) ([]byte, error)
	// DeserializeChangeset encodes a data type from a changeset to its native Go/Kwil type. This can then be used
	// to execute an incoming changeset against a database.
	// TODO: I will have to circle back to actually implement this once I am doing the 2nd half of migrations
	DeserializeChangeset func([]byte) (any, error)
}

var (
	textType = &datatype{
		KwilType:       types.TextType,
		Matches:        []reflect.Type{reflect.TypeOf("")},
		OID:            func(m *pgtype.Map) uint32 { return pgtype.TextOID },
		EncodeInferred: defaultEncodeDecode,
		Decode:         defaultEncodeDecode,
		SerializeChangeset: func(value string) ([]byte, error) {
			return []byte(value), nil
		},
		DeserializeChangeset: func(b []byte) (any, error) {
			return string(b), nil
		},
	}

	textArrayType = &datatype{
		KwilType:       types.TextArrayType,
		Matches:        []reflect.Type{reflect.TypeOf([]string{})},
		OID:            func(m *pgtype.Map) uint32 { return pgtype.TextArrayOID },
		EncodeInferred: defaultEncodeDecode,
		Decode:         decodeArray[string](textType.Decode),
		SerializeChangeset: func(value string) ([]byte, error) {
			// text arrays are delimited by commas, so we need to split on commas.
			// We also need to ensure that the commas
			var ok bool
			value, ok = trimCurlys(value)
			if !ok {
				return nil, fmt.Errorf("invalid text array: %s", value)
			}

			// each string is now wrapped in double quotes in the text literal,
			// e.g. "aaa","bbb","c\"cc"
			// we need to split on "," but not on "\",\""
			inQuote := false
			var strs []string
			currentStr := ""
			for _, char := range value {
				if char == '"' {
					inQuote = !inQuote
				} else if char == ',' && !inQuote {
					strs = append(strs, currentStr)
					currentStr = ""
				} else {
					currentStr += string(char)
				}
			}

			// add the last string
			strs = append(strs, currentStr)

			return serializeArray(strs, 4, textType.SerializeChangeset)
		},
		DeserializeChangeset: deserializeArrayFn[string](4, textType.DeserializeChangeset),
	}

	// we intentionally ignore uint8, since we don't want to cause issues with []byte.
	intType = &datatype{
		KwilType:       types.IntType,
		Matches:        []reflect.Type{reflect.TypeOf(int(0)), reflect.TypeOf(int8(0)), reflect.TypeOf(int16(0)), reflect.TypeOf(int32(0)), reflect.TypeOf(int64(0)), reflect.TypeOf(uint(0)), reflect.TypeOf(uint16(0)), reflect.TypeOf(uint32(0)), reflect.TypeOf(uint64(0))},
		OID:            func(m *pgtype.Map) uint32 { return pgtype.Int8OID },
		ExtraOIDs:      []uint32{pgtype.Int2OID, pgtype.Int4OID},
		EncodeInferred: defaultEncodeDecode,
		Decode: func(a any) (any, error) {
			switch v := a.(type) {
			case int:
				return int64(v), nil
			case int8:
				return int64(v), nil
			case int16:
				return int64(v), nil
			case int32:
				return int64(v), nil
			case int64:
				return v, nil
			case uint:
				return int64(v), nil
			case uint16:
				return int64(v), nil
			case uint32:
				return int64(v), nil
			case uint64:
				return int64(v), nil
			default:
				return nil, fmt.Errorf("unexpected type %T", a)
			}
		},
		SerializeChangeset: func(value string) ([]byte, error) {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}

			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, uint64(intVal))
			return buf, nil
		},
	}

	intArrayType = &datatype{
		KwilType:           types.IntArrayType,
		Matches:            []reflect.Type{reflect.TypeOf([]int{}), reflect.TypeOf([]int8{}), reflect.TypeOf([]int16{}), reflect.TypeOf([]int32{}), reflect.TypeOf([]int64{}), reflect.TypeOf([]uint{}), reflect.TypeOf([]uint16{}), reflect.TypeOf([]uint32{}), reflect.TypeOf([]uint64{})},
		OID:                func(m *pgtype.Map) uint32 { return pgtype.Int8ArrayOID },
		ExtraOIDs:          []uint32{pgtype.Int2ArrayOID, pgtype.Int4ArrayOID},
		EncodeInferred:     defaultEncodeDecode,
		Decode:             decodeArray[int64](intType.Decode),
		SerializeChangeset: arrayFromChildFunc(1, intType.SerializeChangeset),
	}

	boolType = &datatype{
		KwilType:       types.BoolType,
		Matches:        []reflect.Type{reflect.TypeOf(true)},
		OID:            func(m *pgtype.Map) uint32 { return pgtype.BoolOID },
		EncodeInferred: defaultEncodeDecode,
		Decode:         defaultEncodeDecode,
		SerializeChangeset: func(value string) ([]byte, error) {
			if strings.EqualFold(value, "true") || strings.EqualFold(value, "t") {
				return []byte{1}, nil
			}
			if strings.EqualFold(value, "false") || strings.EqualFold(value, "f") {
				return []byte{0}, nil
			}
			return nil, fmt.Errorf("invalid boolean value: %s", value)
		},
	}

	boolArrayType = &datatype{
		KwilType:           types.BoolArrayType,
		Matches:            []reflect.Type{reflect.TypeOf([]bool{})},
		OID:                func(m *pgtype.Map) uint32 { return pgtype.BoolArrayOID },
		EncodeInferred:     defaultEncodeDecode,
		Decode:             decodeArray[bool](boolType.Decode),
		SerializeChangeset: arrayFromChildFunc(1, boolType.SerializeChangeset),
	}

	blobType = &datatype{
		KwilType:       types.BlobType,
		Matches:        []reflect.Type{reflect.TypeOf([]byte{})},
		OID:            func(m *pgtype.Map) uint32 { return pgtype.ByteaOID },
		EncodeInferred: defaultEncodeDecode,
		Decode:         defaultEncodeDecode,
		SerializeChangeset: func(value string) ([]byte, error) {
			// postgres returns all blobs as hex, prefixed with \x
			// we need to remove the \x and decode the hex
			if len(value) < 2 {
				return nil, fmt.Errorf("invalid blob value: %s", value)
			}

			if value[0] != '\\' || value[1] != 'x' {
				return nil, fmt.Errorf("invalid blob value: %s", value)
			}

			return hex.DecodeString(value[2:])
		},
	}

	blobArrayType = &datatype{
		KwilType:           types.BlobArrayType,
		Matches:            []reflect.Type{reflect.TypeOf([][]byte{})},
		OID:                func(m *pgtype.Map) uint32 { return pgtype.ByteaArrayOID },
		EncodeInferred:     defaultEncodeDecode,
		Decode:             decodeArray[[]byte](blobType.Decode),
		SerializeChangeset: arrayFromChildFunc(4, blobType.SerializeChangeset),
	}

	uuidType = &datatype{
		KwilType: types.UUIDType,
		Matches:  []reflect.Type{reflect.TypeOf(types.NewUUIDV5([]byte{})), reflect.TypeOf(*types.NewUUIDV5([]byte{}))},
		OID:      func(m *pgtype.Map) uint32 { return pgtype.UUIDOID },
		EncodeInferred: func(v any) (any, error) {
			var val *types.UUID
			switch v := v.(type) {
			case types.UUID:
				val = &v
			case *types.UUID:
				val = v
			default:
				panic("unreachable")
			}

			return pgtype.UUID{
				Bytes: [16]byte(val.Bytes()),
				Valid: true,
			}, nil
		},
		Decode: func(v any) (any, error) {
			var u types.UUID
			switch v := v.(type) {
			case pgtype.UUID:
				u = types.UUID(v.Bytes)
			case [16]byte:
				u = types.UUID(v)
			default:
				return nil, fmt.Errorf("unexpected type decoding uuid %T", v)
			}
			return &u, nil
		},
		SerializeChangeset: func(value string) ([]byte, error) {
			u, err := types.ParseUUID(value)
			if err != nil {
				return nil, err
			}
			return u.Bytes(), nil
		},
	}

	uuidArrayType = &datatype{
		KwilType: types.UUIDArrayType,
		Matches:  []reflect.Type{reflect.TypeOf(types.UUIDArray{})},
		OID:      func(m *pgtype.Map) uint32 { return pgtype.UUIDArrayOID },
		EncodeInferred: func(v any) (any, error) {
			val, ok := v.(types.UUIDArray)
			if !ok {
				return nil, fmt.Errorf("expected UUIDArray, got %T", v)
			}

			var arr []any
			for _, u := range val {
				v2, err := uuidType.EncodeInferred(u)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v2)
			}

			return arr, nil
		},
		Decode: func(a any) (any, error) {
			arr, ok := a.([]any) // pgx always returns arrays as []any
			if !ok {
				return nil, fmt.Errorf("expected []any, got %T", a)
			}

			vals := make(types.UUIDArray, len(arr))
			for i, v := range arr {
				val, err := uuidType.Decode(v)
				if err != nil {
					return nil, err
				}
				vals[i] = val.(*types.UUID)
			}

			return vals, nil
		},
		SerializeChangeset: arrayFromChildFunc(1, uuidType.SerializeChangeset),
	}

	decimalType = &datatype{
		KwilType: types.DecimalType,
		Matches:  []reflect.Type{reflect.TypeOf(decimal.Decimal{}), reflect.TypeOf(&decimal.Decimal{})},
		OID:      func(m *pgtype.Map) uint32 { return pgtype.NumericOID },
		EncodeInferred: func(v any) (any, error) {
			var dec *decimal.Decimal
			switch v := v.(type) {
			case decimal.Decimal:
				dec = &v
			case *decimal.Decimal:
				dec = v
			default:
				return nil, fmt.Errorf("unexpected type encoding decimal %T", v)
			}

			return pgtype.Numeric{
				Int:   dec.BigInt(),
				Exp:   dec.Exp(),
				Valid: true,
			}, nil
		},
		Decode: func(a any) (any, error) {
			pgType, ok := a.(pgtype.Numeric)
			if !ok {
				return nil, fmt.Errorf("expected pgtype.Numeric, got %T", a)
			}

			if pgType.NaN {
				return "NaN", nil
			}

			// if we give postgres a number such as 5000, it will return it as 5 with exponent 3.
			// Since kwil's decimal semantics do not allow negative scale, we need to multiply
			// the number by 10^exp to get the correct value.
			if pgType.Exp > 0 {
				z := new(big.Int)
				z.Exp(big.NewInt(10), big.NewInt(int64(pgType.Exp)), nil)
				z.Mul(z, pgType.Int)
				pgType.Int = z
				pgType.Exp = 0
			}

			// there is a bit of an edge case here, where uint256 can be returned.
			// since most results simply get returned to the user via JSON, it doesn't
			// matter too much right now, so we'll leave it as-is.
			return decimal.NewFromBigInt(pgType.Int, pgType.Exp)
		},
		SerializeChangeset: func(value string) ([]byte, error) {
			// parse to ensure it is a valid decimal, then re-encode it to ensure it is in the correct format.
			dec, err := decimal.NewFromString(value)
			if err != nil {
				return nil, err
			}

			return []byte(dec.String()), nil
		},
	}

	decimalArrayType = &datatype{
		KwilType: types.DecimalArrayType,
		Matches:  []reflect.Type{reflect.TypeOf(decimal.DecimalArray{})},
		OID:      func(m *pgtype.Map) uint32 { return pgtype.NumericArrayOID },
		EncodeInferred: func(v any) (any, error) {
			val, ok := v.(decimal.DecimalArray)
			if !ok {
				return nil, fmt.Errorf("expected DecimalArray, got %T", v)
			}

			var arr []pgtype.Numeric
			for _, d := range val {
				v2, err := decimalType.EncodeInferred(d)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v2.(pgtype.Numeric))
			}

			return arr, nil
		},
		Decode: func(a any) (any, error) {
			arr, ok := a.([]any) // pgx always returns arrays as []any
			if !ok {
				return nil, fmt.Errorf("expected []any, got %T", a)
			}

			vals := make(decimal.DecimalArray, len(arr))
			for i, v := range arr {
				val, err := decimalType.Decode(v)
				if err != nil {
					return nil, err
				}
				vals[i] = val.(*decimal.Decimal)
			}

			return vals, nil
		},
		SerializeChangeset: arrayFromChildFunc(2, decimalType.SerializeChangeset),
	}

	uint256Type = &datatype{
		KwilType: types.Uint256Type,
		Matches:  []reflect.Type{reflect.TypeOf(types.Uint256{}), reflect.TypeOf(&types.Uint256{})},
		// OID is a custom OID, since Postgres doesn't have a built-in type for uint256,
		// so Kwil uses a Postgres Domain.
		OID: func(m *pgtype.Map) uint32 {
			pgt, ok := m.TypeForName("uint256")
			if !ok {
				// if this happens, it is an internal bug where we are not registering the type
				panic("uint256 domain not found")
			}

			return pgt.OID
		},
		// Under the hood, Kwil's uint256 is a Domain built on a numeric type.
		EncodeInferred: func(a any) (any, error) {
			var val *types.Uint256
			switch v := a.(type) {
			case types.Uint256:
				val = &v
			case *types.Uint256:
				val = v
			default:
				panic("unreachable")
			}

			return pgtype.Numeric{
				Int:   val.ToBig(),
				Exp:   0,
				Valid: true,
			}, nil
		},
		Decode: func(a any) (any, error) {
			pgType, ok := a.(pgtype.Numeric)
			if !ok {
				return nil, fmt.Errorf("expected pgtype.Numeric, got %T", a)
			}

			// if the number ends in 0s, it will have an exponent, so we need to multiply
			// the number by 10^exp to get the correct value.
			if pgType.Exp > 0 {
				z := new(big.Int)
				z.Exp(big.NewInt(10), big.NewInt(int64(pgType.Exp)), nil)
				z.Mul(z, pgType.Int)
				pgType.Int = z
				pgType.Exp = 0
			}

			return types.Uint256FromBig(pgType.Int)
		},
		SerializeChangeset: func(value string) ([]byte, error) {
			// parse to ensure it is a valid uint256, then re-encode it to ensure it is in the correct format.
			u, err := types.Uint256FromString(value)
			if err != nil {
				return nil, err
			}

			return u.Bytes(), nil
		},
	}

	uint256ArrayType = &datatype{
		KwilType: types.Uint256ArrayType,
		Matches:  []reflect.Type{reflect.TypeOf(types.Uint256Array{})},
		// OID is a custom OID, since Postgres doesn't have a built-in type for uint256,
		// See the comment on uint256Type for more information.
		OID: func(m *pgtype.Map) uint32 {
			pgt, ok := m.TypeForName("uint256[]")
			if !ok {
				// if this happens, it is an internal bug where we are not registering the type
				panic("uint256[] domain not found")
			}

			return pgt.OID
		},
		EncodeInferred: func(a any) (any, error) {
			val, ok := a.(types.Uint256Array)
			if !ok {
				return nil, fmt.Errorf("expected Uint256Array, got %T", a)
			}

			vals := make([]pgtype.Numeric, len(val))
			for i, u := range val {
				v2, err := uint256Type.EncodeInferred(u)
				if err != nil {
					return nil, err
				}
				vals[i] = v2.(pgtype.Numeric)
			}

			return vals, nil
		},
		Decode: func(a any) (any, error) {
			arr, ok := a.([]any) // pgx always returns arrays as []any
			if !ok {
				return nil, fmt.Errorf("expected []any, got %T", a)
			}

			vals := make(types.Uint256Array, len(arr))
			for i, v := range arr {
				val, err := uint256Type.Decode(v)
				if err != nil {
					return nil, err
				}
				vals[i] = val.(*types.Uint256)
			}

			return vals, nil
		},
		SerializeChangeset: arrayFromChildFunc(2, uint256Type.SerializeChangeset),
	}
)

// defaultEncodeDecode is the default Encode and Decode function for data types.
// It simply returns the value as is, without any modifications.
func defaultEncodeDecode(v any) (any, error) { return v, nil }

// decodeArrayFn creates a function that decodes an array of a given type.
// it takes a generic for the target scalar type, as well as a decode function
// for the scalar type.
func decodeArray[T any](decode func(any) (any, error)) func(any) (any, error) {
	return func(a any) (any, error) {
		arr, ok := a.([]any) // pgx always returns arrays as []any
		if !ok {
			return nil, fmt.Errorf("expected []any, got %T", a)
		}

		vals := make([]T, len(arr))
		for i, v := range arr {
			val, err := decode(v)
			if err != nil {
				return nil, err
			}
			vals[i] = val.(T)
		}

		return vals, nil
	}
}

// encodeToPGType encodes several Go types to their corresponding pgx types.
// It is capable of detecting special Kwil types and encoding them to their
// corresponding pgx types. It is only used if using inferred argument types.
// If not using inferred argument types, pgx will rely on the Valuer interface
// to encode the Go types to their corresponding pgx types.
// It also returns the pgx type OIDs for each value.
func encodeToPGType(oids *pgtype.Map, values ...any) ([]any, []uint32, error) {
	if len(values) == 0 {
		return nil, nil, nil
	}

	encoded := make([]any, len(values))
	oidsArr := make([]uint32, len(values))
	for i, v := range values {
		if v == nil {
			encoded[i] = nil
			oidsArr[i] = pgtype.TextOID
			continue
		}

		// special case, if []any, we need to encode each element
		if arr, ok := v.([]any); ok {
			if len(arr) == 0 {
				encoded[i] = nil
				oidsArr[i] = pgtype.TextOID
				continue
			}

			encodedArr, oidsArrArr, err := encodeToPGType(oids, arr...)
			if err != nil {
				return nil, nil, err
			}

			encoded[i] = encodedArr

			// check that all OIDs are the same
			oid := oidsArrArr[0]
			for _, oid2 := range oidsArrArr {
				if oid != oid2 {
					return nil, nil, fmt.Errorf("all elements in an array must have the same data type")
				}
			}

			dt, ok := dataTypesByMatch[reflect.TypeOf(arr[0])]
			if !ok {
				return nil, nil, fmt.Errorf("unsupported type %T", arr[0])
			}

			arrDt, ok := scalarToArray[dt]
			if !ok {
				return nil, nil, fmt.Errorf("no array type for %T", arr[0])
			}

			oidsArr[i] = arrDt.OID(oids)

			continue
		}

		dt, ok := dataTypesByMatch[reflect.TypeOf(v)]
		if !ok {
			return nil, nil, fmt.Errorf("unsupported type %T", v)
		}

		encodedVal, err := dt.EncodeInferred(v)
		if err != nil {
			return nil, nil, err
		}

		encoded[i] = encodedVal
		oidsArr[i] = dt.OID(oids)
	}

	return encoded, oidsArr, nil
}

// for functions that return void, it will actually return
// a nil value with the void OID.
var voidOID = uint32(2278)

// decodeFromPGType decodes several pgx types to their corresponding Go types.
// It is capable of detecting special Kwil types and decoding them to their
// corresponding Go types.
func decodeFromPG(vals []any, oids []uint32, oidToDataType map[uint32]*datatype) ([]any, error) {
	var results []any
	for i, oid := range oids {
		if oid == voidOID {
			continue
		}

		if vals[i] == nil {
			results = append(results, nil)
			continue
		}

		dt, ok := oidToDataType[oid]
		if !ok {
			return nil, fmt.Errorf("unsupported oid %d", oid)
		}

		decoded, err := dt.Decode(vals[i])
		if err != nil {
			return nil, err
		}

		results = append(results, decoded)
	}

	return results, nil
}

// oidTypesMap makes a map mapping oids to the Kwil type definition.
// It needs to be called after registerTypes.
func oidTypesMap(conn *pgtype.Map) map[uint32]*datatype {
	m := make(map[uint32]*datatype)
	for dt := range datatypes {
		oid := dt.OID(conn)
		_, ok := m[oid]
		if ok {
			panic("duplicate oid for type. OID:" + fmt.Sprint(oid))
		}
		m[oid] = dt

		for _, extraOID := range dt.ExtraOIDs {
			_, ok := m[extraOID]
			if ok {
				panic("duplicate oid for type. OID:" + fmt.Sprint(extraOID))
			}
			m[extraOID] = dt
		}
	}
	return m
}

// trimCurlys parses curly brackets on the outside of a string.
// It returns the string without the curly brackets, and a boolean
// indicating whether the string had curly brackets. It is useful
// for parsing stringified Postgres arrays.
func trimCurlys(s string) (string, bool) {
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return s[1 : len(s)-1], true
	}

	return s, false
}

// serializeArray serializes an array of some type to []byte.
// It takes a function that serializes the scalar values to []byte.
// lengthSize is the byte size of the length of each element, which allows
// us to more efficiently serialize arrays of fixed-size elements (int, bool, etc).
// lengthSize must be 1, 2, or 4, corresponding to 8-bit, 16-bit, and 32-bit lengths.
func serializeArray[T any](arr []T, lengthSize uint8, serialize func(T) ([]byte, error)) ([]byte, error) {
	encodeLength := func(length int) []byte {
		switch lengthSize {
		case 1:
			return []byte{byte(length)}
		case 2:
			buf := make([]byte, 2)
			binary.BigEndian.PutUint16(buf, uint16(length))
			return buf
		case 4:
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, uint32(length))
			return buf
		default:
			panic("invalid length size")
		}
	}

	var buf []byte
	for _, v := range arr {
		encoded, err := serialize(v)
		if err != nil {
			return nil, err
		}

		buf = append(buf, encodeLength(len(encoded))...)
		buf = append(buf, encoded...)
	}

	return buf, nil
}

// deserializeArray deserializes an array of some type from []byte.
// It takes a function that deserializes the scalar values from []byte.
// it is the inverse of serializeArray. lengthSize must be 1, 2, or 4,
// corresponding to 8-bit, 16-bit, and 32-bit lengths.
func deserializeArray[T any](buf []byte, lengthSize uint8, deserialize func([]byte) (any, error)) ([]T, error) {
	// the lengthSize thing might be a bit overkill, but it is very encapsulated so
	// I'll keep it for now, since it can help decrease the size of the changeset that
	// a network has to process.
	determineLength := func(buf []byte) (int, []byte) {
		switch lengthSize {
		case 1:
			return int(buf[0]), buf[1:]
		case 2:
			return int(binary.BigEndian.Uint16(buf[:2])), buf[2:]
		case 4:
			return int(binary.BigEndian.Uint32(buf[:4])), buf[4:]
		default:
			panic("invalid length size")
		}
	}

	var arr []T
	for len(buf) > 0 {
		length, rest := determineLength(buf)

		v, err := deserialize(rest[:length])
		if err != nil {
			return nil, err
		}

		arr = append(arr, v.(T))
		buf = rest[length:]
	}

	return arr, nil
}

// arrayFromChildFunc splits a stringified array into its elements, and uses
// the callback function to serialize each element. It is meant to be used with
// array data types that do not have special parsing rules. It returns it as a function
// that can be used for decoding changesets
func arrayFromChildFunc(size uint8, serialize func(string) ([]byte, error)) func(string) ([]byte, error) {
	return func(s string) ([]byte, error) {
		s, ok := trimCurlys(s)
		if !ok {
			return nil, fmt.Errorf("invalid array: %s", s)
		}

		strs := strings.Split(s, ",")
		return serializeArray(strs, size, serialize)
	}
}

// deserializeArrayFn returns a function that deserializes an array of some type from a serialized array.
// It is the logical inverse of arrayFromChildFunc.
func deserializeArrayFn[T any](size uint8, deserialize func([]byte) (any, error)) func([]byte) (any, error) {
	return func(b []byte) (any, error) {
		return deserializeArray[T](b, size, deserialize)
	}
}
