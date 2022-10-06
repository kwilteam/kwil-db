package messaging

type Message interface {
}

type RawMessage struct {
	Key   []byte
	Value []byte
}
