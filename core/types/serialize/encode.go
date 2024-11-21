package serialize

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/rlp"
)

// EncodingType is the type used to enumerate different codecs for binary data.
type EncodingType = uint16

const EncodingTypeCustom = 1 << 12 // 4096 codecs reserved for kwild

const (
	// it is very important that the order of the encoding types is not changed
	encodingTypeInvalid EncodingType = iota
	encodingTypeRLP
)

// Codec contains the encoding and decoding functionality for a certain
// serialization scheme that may be used in the Encode, Decode, and Decode
// methods for payloads with the matching EncodingType.
type Codec struct {
	Type   EncodingType
	Name   string
	Encode func(any) ([]byte, error)
	Decode func([]byte, any) error
}

var (
	rlpCodec = Codec{
		Type: encodingTypeRLP,
		Name: "RLP",
		Encode: func(val any) ([]byte, error) {
			return rlp.EncodeToBytes(val)
		},
		Decode: func(bts []byte, v any) error {
			return rlp.DecodeBytes(bts, v)
		},
	}
)

var encodings = map[EncodingType]Codec{
	encodingTypeRLP: rlpCodec,
}

// RegisterCodec installs a new external codec. The codec extension
// implementation should choose a Type that does not collide with other codecs.
// The EncodingTypeCustom offset should be used as the first possible external
// codec's type to leave space for more kwild canonical codecs in the future.
//
// core cannot require main module, only reverse, so registry is here, and
// extensions/consensus.RegisterCodec is provided to ensure the same registry
// used by kwild is used when extension authors define a new codec.
func RegisterCodec(c *Codec) {
	encType := c.Type
	if encType <= EncodingTypeCustom {
		panic(fmt.Sprintf("reserved codec type %d", encType))
	}
	if c0, have := encodings[encType]; have {
		panic(fmt.Sprintf("already have codec %d (%v)", encType, c0.Name))
	}
	encodings[encType] = *c
}

// Encode encodes the given value. If the value is an encoding.BinaryMarshaler,
// its MarshalBinary method is used, otherwise it uses this package's current
// serialized data format (RLP).
func Encode(val any) ([]byte, error) {
	return EncodeWithCodec(val, rlpCodec)
}

// EncodeWithCodec encodes the given value into a serialized data format with
// the provided Codec.
func EncodeWithCodec(val any, enc Codec) ([]byte, error) {
	btsVal, err := enc.Encode(val)
	if err != nil {
		return nil, err
	}

	return addSerializedTypePrefix(enc.Type, btsVal), nil
}

func requireNonNilPointer(v any) error {
	rVal := reflect.ValueOf(v)
	if rType := rVal.Type(); rType.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer: %s / %T", rType.Kind(), v)
	}
	if rVal.IsNil() {
		return errors.New("cannot decode into nil pointer")
	}
	return nil
}

// Decode decodes the data into a value, which should be passed as a pointer. If
// the value is an encoding.BinaryUnmarshaler, its UnmarshalBinary method is
// used, otherwise it will attempt to decode the data using as if it were
// encoded with EncodeWithCodec (checking for a serialized type prefix).
func Decode(bts []byte, v any) error {
	if err := requireNonNilPointer(v); err != nil {
		return err
	}

	encType, val, err := removeSerializedTypePrefix(bts)
	if err != nil {
		return err
	}

	codec, have := encodings[encType]
	if !have {
		return fmt.Errorf("unknown encoding type %v", encType)
	}

	return codec.Decode(val, v)
}

// DecodeGeneric decodes the given serialized data into the given value. See
// also Decode for use with an existing instance. This is generally no more
// useful than Decode; it is syntactic sugar that requires no existing instance
// of the type, and returns a pointer to the declared type.
func DecodeGeneric[T any](bts []byte) (*T, error) {
	var val T
	if err := Decode(bts, &val); err != nil {
		return nil, err
	}
	return &val, nil
}

// TODO: probably remove all the below slice stuff if we don't use it

func EncodeSlice[T any](kvs []T) ([]byte, error) {
	marshaller := make([]*serialBinaryMarshaller[T], len(kvs))
	for i, kv := range kvs {
		marshaller[i] = &serialBinaryMarshaller[T]{kv}
	}
	return serializeSlice[*serialBinaryMarshaller[T]](marshaller)
}

func DecodeSlice[T any](bts []byte) ([]*T, error) {
	marshaller, err := deserializeSlice[*serialBinaryMarshaller[T]](bts, func() *serialBinaryMarshaller[T] {
		return &serialBinaryMarshaller[T]{}
	})
	if err != nil {
		return nil, err
	}

	result := make([]*T, len(marshaller))
	for i, m := range marshaller {
		result[i] = &m.val
	}

	return result, nil
}

// serialBinaryMarshaller is a helper struct that implements the BinaryMarshaler and BinaryUnmarshaler interfaces
type serialBinaryMarshaller[T any] struct {
	val T
}

func (m *serialBinaryMarshaller[T]) MarshalBinary() ([]byte, error) {
	return Encode(m.val)
}

func (m *serialBinaryMarshaller[T]) UnmarshalBinary(bts []byte) error {
	return Decode(bts, &m.val)
}
