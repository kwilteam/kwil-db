package rlp

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

type encodingType uint16

const (
	// it is very important that the order of the encoding types is not changed
	encodingTypeInvalid encodingType = iota
	encodingTypeRLP
)

var currentEncodingType = encodingTypeRLP

func Encode(val any) ([]byte, error) {
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

	return serializeEncodedValue(&encodedValue{
		EncodingType: currentEncodingType,
		Data:         btsVal,
	})
}

func Decode[T any](bts []byte) (*T, error) {
	encVal, err := deserializeEncodedValue(bts)
	if err != nil {
		return nil, err
	}

	switch encVal.EncodingType {
	case encodingTypeRLP:
		return decodeRLP[T](encVal.Data)
	default:
		return nil, fmt.Errorf("invalid encoding type: %d", encVal.EncodingType)
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
