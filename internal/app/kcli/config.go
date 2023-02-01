package kcli

import (
	fund2 "kwil/pkg/fund"
)

type Config struct {
	// GRPC client config
	Endpoint string `json:"endpoint" yaml:"endpoint" toml:"endpoint" mapstructure:"endpoint"`
	// Fund config
	Fund *fund2.Config `json:"fund" yaml:"fund" toml:"fund" mapstructure:"fund"`
}
