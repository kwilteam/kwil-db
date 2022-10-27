package sub

import (
	"kwil/x"
)

type ChannelBroker interface {
	Start() error
	Stop()
	OnStop() <-chan x.Void

	OnChannelAssigned() <-chan ReceiverChannel
}
