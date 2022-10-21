package sub

import (
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
)

type MessageIterator[T any] interface {
	HasNext() bool
	Next() (T, mx.Offset)

	// Commit is used to signal the broker to commit the
	// largest offset consumed for the batch. The
	// offset committed will be based on how far the
	// message iterator has been advanced.
	Commit() async.Action
}
