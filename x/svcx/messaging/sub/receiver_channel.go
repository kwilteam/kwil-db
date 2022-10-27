package sub

import (
    "kwil/x"
    "kwil/x/svcx/messaging/mx"
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
