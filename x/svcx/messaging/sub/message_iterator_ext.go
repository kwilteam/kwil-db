package sub

import (
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
	"math"
)

type message_iterator struct {
	partitionId mx.PartitionId
	next        func() (msg *mx.RawMessage, offset mx.Offset, ok bool)
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

	msg, o, ok := m.next()
	if !ok {
		m.message = nil
		m.offset = math.MinInt
	} else {
		m.message = msg
		m.offset = o
	}
	return ok
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
