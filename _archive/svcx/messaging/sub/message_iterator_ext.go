package sub

import (
	"kwil/archive/svcx/messaging/mx"
	"kwil/x/async"
	"math"
)

type message_iterator struct {
	partitionId mx.PartitionId
	next        func() (msg *mx.RawMessage, offset mx.Offset)
	commit      func() async.Action
	message     *mx.RawMessage
	offset      mx.Offset
}

func (m *message_iterator) PartitionId() mx.PartitionId {
	return m.partitionId
}

func (m *message_iterator) HasNext() bool {
	if m.offset == math.MinInt {
		return false
	}

	m.message, m.offset = m.next()

	return m.offset != math.MinInt
}

func (m *message_iterator) Next() (*mx.RawMessage, mx.Offset) {
	if m.offset == math.MinInt {
		panic("no more messages")
	}

	return m.message, m.offset
}

func (m *message_iterator) Commit() async.Action {
	return m.commit()
}
