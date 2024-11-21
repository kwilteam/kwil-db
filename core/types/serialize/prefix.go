package serialize

import (
	"encoding/binary"
	"errors"
)

// addSerializedTypePrefix adds a prefix to the encoded value to indicate the encoding type.
func addSerializedTypePrefix(encoding EncodingType, encodedValue []byte) []byte {
	encodingTypeBytes := uint16ToBytes(encoding)
	return append(encodingTypeBytes, encodedValue...)
}

// removeSerializedTypePrefix removes the prefix from the encoded value.
func removeSerializedTypePrefix(data []byte) (EncodingType, []byte, error) {
	if len(data) < 3 {
		return encodingTypeInvalid, nil, errors.New("cannot deserialize encoded value: data is too short")
	}
	typ, err := bytesToUint16(data[:2])
	if err != nil {
		return encodingTypeInvalid, nil, err
	}

	return typ, data[2:], nil
}

// uint16ToBytes converts a uint16 to a byte slice (big endian).
func uint16ToBytes(n uint16) []byte {
	return binary.BigEndian.AppendUint16(nil, n)
}

// bytesToUint16 converts a byte slice to a uint16.
func bytesToUint16(b []byte) (uint16, error) {
	if len(b) != 2 {
		return 0, errors.New("cannot convert bytes to uint16: incorrect length")
	}
	return binary.BigEndian.Uint16(b), nil
}
