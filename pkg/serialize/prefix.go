package serialize

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// addSerializedTypePrefix adds a prefix to the encoded value to indicate the encoding type.
func addSerializedTypePrefix(encoding encodingType, encodedValue []byte) (SerializedData, error) {
	result, err := uint16ToBytes(uint16(encoding))
	if err != nil {
		return nil, err
	}

	return append(result, encodedValue...), nil
}

// removeSerializedTypePrefix removes the prefix from the encoded value.
func removeSerializedTypePrefix(data SerializedData) (encodingType, []byte, error) {
	if len(data) == 0 {
		return encodingTypeInvalid, nil, fmt.Errorf("cannot deserialize encoded value: data is empty")
	}
	typ, err := bytesToUint16(data[:2])
	if err != nil {
		return encodingTypeInvalid, nil, err
	}

	return encodingType(typ), data[2:], nil
}

// uint16ToBytes converts a uint16 to a byte slice.
func uint16ToBytes(n uint16) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, n)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// bytesToUint16 converts a byte slice to a uint16.
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
