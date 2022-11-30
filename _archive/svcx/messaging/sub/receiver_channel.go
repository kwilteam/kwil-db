package sub

import (
	"kwil/_archive/svcx/messaging/mx"
	"kwil/x"
)

//var ErrReceiverClosed = fmt.Errorf("receiver closed")
//var ErrUnexpectedReceiverError = fmt.Errorf("receiver event response unknown")

type ReceiverChannel interface {
	Topic() string
	PartitionId() mx.PartitionId

	OnReceive() <-chan MessageIterator

	Stop()
	OnStop() <-chan x.Void
}
