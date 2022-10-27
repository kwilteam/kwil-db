package sub

import (
    "kwil/x"
)

type TransientReceiver interface {
    Topic() string

    OnReceive() <-chan MessageIterator

    Start() error
    Stop()
    OnStop() <-chan x.Void
}
