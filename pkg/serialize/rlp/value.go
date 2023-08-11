package rlp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type encodedValue struct {
	EncodingType encodingType
	Data         []byte
}

func serializeEncodedValue(value *encodedValue) ([]byte, error) {
	result, err := uint16ToBytes(uint16(value.EncodingType))
	if err != nil {
		return nil, err
	}

	return append(result, value.Data...), nil
}

func deserializeEncodedValue(data []byte) (*encodedValue, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot deserialize encoded value: data is empty")
	}
	typ, err := bytesToUint16(data[:2])
	if err != nil {
		return nil, err
	}

	return &encodedValue{
		EncodingType: encodingType(typ),
		Data:         data[2:],
	}, nil
}

func uint16ToBytes(n uint16) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, n)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func bytesToUint16(b []byte) (uint16, error) {
	if len(b) < 2 {
		return 0, fmt.Errorf("cannot convert bytes to uint16: bytes are too short")
	}

	buf := bytes.NewReader(b)
	var n uint16
	err := binary.Read(buf, binary.BigEndian, &n)
	if err != nil {
		return 0, err
	}
	return n, nil
}
