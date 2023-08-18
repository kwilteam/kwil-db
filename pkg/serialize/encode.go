package serialize

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

type SerializedData = []byte

type encodingType uint16

const (
	// it is very important that the order of the encoding types is not changed
	encodingTypeInvalid encodingType = iota
	encodingTypeRLP
)

var currentEncodingType = encodingTypeRLP

func Encode(val any) (SerializedData, error) {
	var btsVal []byte
	var err error
	switch currentEncodingType {
	case encodingTypeRLP:
		btsVal, err = encodeRLP(val)
	default:
		return nil, fmt.Errorf("invalid encoding type: %d", currentEncodingType)
	}
	if err != nil {
		return nil, err
	}

	return addSerializedTypePrefix(currentEncodingType, btsVal)
}

func Decode[T any](bts SerializedData) (*T, error) {
	encType, val, err := removeSerializedTypePrefix(bts)
	if err != nil {
		return nil, err
	}

	switch encType {
	case encodingTypeRLP:
		return decodeRLP[T](val)
	default:
		return nil, fmt.Errorf("invalid encoding type: %d", val)
	}
}

func encodeRLP(val any) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

func decodeRLP[T any](bts []byte) (*T, error) {
	var val T
	err := rlp.DecodeBytes(bts, &val)
	if err != nil {
		return nil, err
	}

	return &val, nil
}

func DecodeInto(bts []byte, v any) error {
	encType, val, err := removeSerializedTypePrefix(bts)
	if err != nil {
		return err
	}

	switch encType {
	case encodingTypeRLP:
		return rlp.DecodeBytes(val, v)
	default:
		return fmt.Errorf("invalid encoding type: %d", val)
	}
}
