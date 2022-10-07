package messaging

import "kwil/x/rx"

type AckNackFn func(err error) rx.Continuation

var none_continuation = rx.SuccessC()

var NoAckNack AckNackFn = func(err error) rx.Continuation {
	return none_continuation
}

type RawMessage struct {
	Key   []byte
	Value []byte
}

type ProducerMessage[T any] interface {
	Payload() T
	GetAckNack() AckNackFn // optional
}

type ConsumerBatch[T any] interface {
	Source() string
	Grouped() []Group[T]
}

type ConsumerMessage[T any] interface {
	Id() int64
	Payload() T
}

type Group[T any] interface {
	Id() int32
	Messages() ConsumerMessage[T]

	GetAckNack() AckNackFn // optional
}

func MessageNoAckP[T any](payload T) ProducerMessage[T] {
	return &producer_message_no_ack[T]{payload}
}

func MessageP[T any](payload T, ack AckNackFn) ProducerMessage[T] {
	if ack == nil {
		return &producer_message_no_ack[T]{payload}
	}

	return &producer_message[T]{payload, ack}
}

type producer_message_no_ack[T any] struct {
	payload T
}

func (m *producer_message_no_ack[T]) Payload() T {
	return m.payload
}

func (m *producer_message_no_ack[T]) GetAckNack() AckNackFn {
	return NoAckNack
}

type producer_message[T any] struct {
	payload T
	ack     AckNackFn
}

func (m *producer_message[T]) Payload() T {
	return m.payload
}

func (m *producer_message[T]) GetAckNack() AckNackFn {
	return m.ack
}

func AckNack(fn func(error)) AckNackFn {
	return func(err error) rx.Continuation {
		fn(err)
		return none_continuation
	}
}
