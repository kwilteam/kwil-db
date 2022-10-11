package sub

import (
	"context"
	"kwil/x"
)

type ChannelBroker[T any] interface {
	Start(topics ...string) error

	OnChannelAssigned() <-chan ReceiverChannel[T]

	Shutdown()
	ShutdownAndWait(ctx context.Context) error
	OnShutdown() <-chan x.Void
}
