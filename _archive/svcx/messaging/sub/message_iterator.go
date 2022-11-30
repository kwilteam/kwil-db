package sub

import (
	"kwil/_archive/svcx/messaging/mx"
	"kwil/x/async"
)

type MessageIterator interface {
	PartitionId() mx.PartitionId

	HasNext() bool
	Next() (*mx.RawMessage, mx.Offset)

	// Commit is used to signal the broker to commit the
	// largest offset consumed for the batch. The
	// offset committed will be based on how far the
	// message iterator has been advanced.
	Commit() async.Action
}
