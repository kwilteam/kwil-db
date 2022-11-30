package wallet

import (
	"fmt"
	"kwil/_archive/svcx/messaging/mx"
)

type ConfirmationEvent struct {
	request_id string
	message    *mx.RawMessage
}

func encode_event(request_id string, msg *mx.RawMessage) *mx.RawMessage {
	payload := []byte(request_id)
	return &mx.RawMessage{Key: msg.Key, Value: append(payload, msg.Value...)}
}

func decode_event(msg *mx.RawMessage) (*mx.RawMessage, string, error) {
	if len(msg.Value) == 0 {
		return nil, "", fmt.Errorf("empty message")
	}

	if len(msg.Value) < 36 {
		return nil, "", fmt.Errorf("invalid request message")
	}

	request_id := string(msg.Value[0:36])
	payload := msg.Value[36:]

	return &mx.RawMessage{Key: msg.Key, Value: payload}, request_id, nil
}
