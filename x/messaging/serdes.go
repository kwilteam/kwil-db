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
	Serialize(T) (key []byte, value []byte, err error)

	// Deserialize deserializes a message from a byte array.
	Deserialize(key []byte, value []byte) (T, error)
}
