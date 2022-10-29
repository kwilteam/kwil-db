package wallet

import (
	"fmt"
	"github.com/google/uuid"
	"kwil/x/svcx/messaging/mx"
)

type Request struct {
	WalletId string // used as key
	Payload  []byte
}

func encode_message_async(msg *mx.RawMessage) *mx.RawMessage {
	payload := []byte{byte(0)}
	return &mx.RawMessage{Key: msg.Key, Value: append(payload, msg.Value...)}
}

func encode_message(msg *mx.RawMessage) (*mx.RawMessage, string) {
	request_id := uuid.New().String()

	payload := []byte{byte(1)}
	payload = append(payload, []byte(request_id)...)

	return &mx.RawMessage{Key: msg.Key, Value: append(payload, msg.Value...)}, request_id
}

func decode_message(msg *mx.RawMessage) (*mx.RawMessage, string, error) {
	if len(msg.Value) == 0 {
		return nil, "", fmt.Errorf("empty message")
	}

	if msg.Value[0] == byte(0) {
		return &mx.RawMessage{Key: msg.Key, Value: msg.Value[1:]}, "", nil
	}

	if len(msg.Value) < 37 {
		return nil, "", fmt.Errorf("invalid request message")
	}

	request_id := string(msg.Value[1:37])
	payload := msg.Value[37:]

	return &mx.RawMessage{Key: msg.Key, Value: payload}, request_id, nil
}
