package mx

import (
	cfg "kwil/x/cfgx"
)

type Offset int64
type PartitionId int32

type ChannelInfo interface {
	Topic() string
	PartitionId() PartitionId
}

type ChannelConfig[T any] struct {
	config cfg.Config
	serdes Serdes[T]
}

func (c *ChannelConfig[T]) Config() cfg.Config {
	return c.config
}

func (c *ChannelConfig[T]) Serdes() Serdes[T] {
	return c.serdes
}

func NewChannelConfig[T any](Config cfg.Config, Serdes Serdes[T]) *ChannelConfig[T] {
	return &ChannelConfig[T]{Config, Serdes}
}
