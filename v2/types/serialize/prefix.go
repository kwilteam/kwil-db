package serialize

import (
	"encoding/binary"
	"fmt"
)

// addSerializedTypePrefix adds a prefix to the encoded value to indicate the encoding type.
func addSerializedTypePrefix(encoding EncodingType, encodedValue []byte) SerializedData {
	encodingTypeBytes := uint16ToBytes(encoding)
	return append(encodingTypeBytes, encodedValue...)
}

// removeSerializedTypePrefix removes the prefix from the encoded value.
func removeSerializedTypePrefix(data SerializedData) (EncodingType, []byte, error) {
	if len(data) < 3 {
		return encodingTypeInvalid, nil, fmt.Errorf("cannot deserialize encoded value: data is too short")
	}
	typ, err := bytesToUint16(data[:2])
	if err != nil {
		return encodingTypeInvalid, nil, err
	}

	return typ, data[2:], nil
}

// uint16ToBytes converts a uint16 to a byte slice (big endian).
func uint16ToBytes(n uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	return b
}

// bytesToUint16 converts a byte slice to a uint16.
func bytesToUint16(b []byte) (uint16, error) {
	if len(b) != 2 {
		return 0, fmt.Errorf("cannot convert bytes to uint16: incorrect length")
	}
	return binary.BigEndian.Uint16(b), nil
}
