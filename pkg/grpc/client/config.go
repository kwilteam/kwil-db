package client

import (
	"go.uber.org/zap"
)

type GrpcConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

func (c *GrpcConfig) toConfig() (*clientConfig, error) {
	return &clientConfig{
		Endpoint: c.Endpoint,
	}, nil
}

type clientConfig struct {
	Endpoint string
	// TODO: use logger
	Log *zap.Logger
}
