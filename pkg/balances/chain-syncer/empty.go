package chainsyncer

import "context"

type EmptyChainsyncer struct{}

func (c *EmptyChainsyncer) Start(ctx context.Context) error {
	return nil
}

func NewEmptyChainSyncer() *EmptyChainsyncer {
	return &EmptyChainsyncer{}
}
