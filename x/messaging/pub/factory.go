package pub

import (
	"fmt"
	"kwil/x/cfgx"
	"kwil/x/messaging/mx"
)

func NewEmitter[T any](config cfgx.Config, serdes mx.Serdes[T]) (Emitter[T], error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if serdes == nil {
		return nil, fmt.Errorf("serdes is nil")
	}

	cfg, err := mx.NewEmitterConfig[T](config, serdes)
	if err != nil {
		return nil, err
	}

	return start[T](cfg)
}

func NewMessageNoAck[T any](topic string, payload T) Message[T] {
	return &producer_message_no_ack[T]{topic, payload}
}

func NewMessage[T any](topic string, payload T, ack AckNackFn) Message[T] {
	if ack == nil {
		return &producer_message_no_ack[T]{topic, payload}
	}

	return &producer_message[T]{topic, payload, ack}
}
