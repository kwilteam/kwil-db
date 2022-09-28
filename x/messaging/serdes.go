package messaging

type Serdes[T any] interface {
	Serialize(T) (key []byte, value []byte, err error)
	Deserialize(key []byte, value []byte) (T, error)
}
