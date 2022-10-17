package mx

type RawMessage struct {
	Key   []byte
	Value []byte
}

var emptyRawMessage = RawMessage{}

// EmptyRawMessage returns an empty RawMessage.
func EmptyRawMessage() RawMessage {
	return emptyRawMessage
}
