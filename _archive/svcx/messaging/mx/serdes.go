package mx

// Serdes is the primary way by which we will convert
// between the application's domain model and the byte
// array expected by most composer layers. This will
// ultimately include the ability to convert to and
// from JSON, Avro, and Protobuf. In addition, it will
// include the ability to use a schema registry for
// backward/forward compatibility.
type Serdes[T any] interface {
	// Serialize serializes T into a byte array.
	Serialize(data T) (key []byte, value []byte, err error)

	// Deserialize deserializes T from a byte array.
	Deserialize(key []byte, value []byte) (T, error)
}
