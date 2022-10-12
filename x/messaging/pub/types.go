package pub

import (
	"fmt"
	"kwil/x/rx"
)

var errEmitterNotFound = fmt.Errorf("emitter not found")

func ErrEmitterNotFound() error {
	return errEmitterNotFound
}

type AckNackFn func(err error) rx.Action

var none_ack AckNackFn = func(err error) rx.Action {
	return none_action
}

var none_action = rx.SuccessA()

type Message[T any] struct {
	Topic   string // optional if using default-topic
	Payload T
	AckNack AckNackFn // optional
}

func (m *Message[T]) ack_nack(err error) rx.Action {
	if m.AckNack == nil {
		return none_action
	}
	return m.AckNack(err)
}

func AckNackSync(fn func(error)) AckNackFn {
	return func(err error) rx.Action {
		fn(err)
		return none_action
	}
}

func NewMessage[T any](payload T, ackNack AckNackFn) *Message[T] {
	return &Message[T]{"", payload, ackNack}
}

func NewMessageNoAck[T any](payload T) *Message[T] {
	return &Message[T]{Payload: payload}
}

func NewMessageT[T any](topic string, payload T, ackNack AckNackFn) *Message[T] {
	return &Message[T]{topic, payload, ackNack}
}

func NewMessageNoAckT[T any](topic string, payload T) *Message[T] {
	return &Message[T]{topic, payload, nil}
}
