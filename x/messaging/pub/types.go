package pub

import (
	"kwil/x/rx"
)

type AckNackFn func(err error) rx.Action

var none_action = rx.SuccessA()

var NoAckNack AckNackFn = func(err error) rx.Action {
	return none_action
}

type Message[T any] interface {
	Topic() string
	Payload() T
	GetAckNack() AckNackFn // optional
}

type producer_message_no_ack[T any] struct {
	topic   string
	payload T
}

func (m *producer_message_no_ack[T]) Topic() string {
	return m.topic
}

func (m *producer_message_no_ack[T]) Payload() T {
	return m.payload
}

func (m *producer_message_no_ack[T]) GetAckNack() AckNackFn {
	return NoAckNack
}

type producer_message[T any] struct {
	topic   string
	payload T
	ack     AckNackFn
}

func (m *producer_message[T]) Topic() string {
	return m.topic
}

func (m *producer_message[T]) Payload() T {
	return m.payload
}

func (m *producer_message[T]) GetAckNack() AckNackFn {
	return m.ack
}

func AckNack(fn func(error)) AckNackFn {
	return func(err error) rx.Action {
		fn(err)
		return none_action
	}
}
