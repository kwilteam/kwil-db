package mx

var serdesByteArray Serdes[RawMessage] = &serdes_byte_array{}

// SerdesByteArray converts back and forth to array.
// Key is ignored.
func SerdesByteArray() Serdes[RawMessage] {
	return serdesByteArray
}

type serdes_byte_array struct{}

func (_ *serdes_byte_array) Serialize(m RawMessage) ([]byte, []byte, error) {
	return m.Key, m.Value, nil
}

func (_ *serdes_byte_array) Deserialize(key, value []byte) (RawMessage, error) {
	return RawMessage{key, value}, nil
}
