package messaging

// Serdes is the primary way by which we will convert
// between the application's domain model and the byte
// array expected by kafka. This will ultimately include
// the ability to convert to and from JSON, Avro, and
// Protobuf. In addition, it will include the ability
// to use a schema registry for backward/forward
// compatibility.
type Serdes[T any] interface {
	// Serialize serializes a message into a byte array.
	Serialize(message T) (key []byte, value []byte, err error)

	// Deserialize deserializes a message from a byte array.
	Deserialize(key []byte, value []byte) (T, error)
}

// SerdesByteArray converts value back and forth to array.
// Key is ignored.
var SerdesByteArray Serdes[RawMessage] = &serdes_byte_array{}

type serdes_byte_array struct{}

func (_ *serdes_byte_array) Serialize(m RawMessage) ([]byte, []byte, error) {
	return m.Key, m.Value, nil
}

func (_ *serdes_byte_array) Deserialize(key, value []byte) (RawMessage, error) {
	return RawMessage{key, value}, nil
}
