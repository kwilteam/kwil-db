package kcli

import (
	"kwil/pkg/fund"
	"kwil/pkg/grpc/client"
)

type Config struct {
	// Kwil config
	Kwil client.GrpcConfig `json:"kwil" yaml:"kwil" toml:"kwil" mapstructure:"kwil"`
	// Fund config
	Fund fund.Config `json:"fund" yaml:"fund" toml:"fund" mapstructure:"fund" env`
}
