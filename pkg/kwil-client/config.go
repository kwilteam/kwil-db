package kwil_client

import (
	"kwil/pkg/fund"
	"kwil/pkg/grpc/client"
)

type Config struct {
	// Node config
	// @yaiba TODO: a better name, maybe Peer?
	Node client.GrpcConfig `mapstructure:"node"`
	// Fund config
	// @yaiba TODO: a better name, maybe SettlementChain?
	Fund fund.Config `mapstructure:"fund"`
}
