package client

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
)

type KwildConfig struct {
	PrivateKey        string
	GrpcURL           string
	ClientChainRPCURL string
}

func NewClient(ctx context.Context, cfg *KwildConfig) (*client.Client, error) {
	options := []client.ClientOpt{}
	if cfg.ClientChainRPCURL != "" {
		options = append(options, client.WithCometBftUrl(cfg.ClientChainRPCURL))
	}
	if cfg.PrivateKey != "" {
		key, _ := crypto.ECDSAFromHex(cfg.PrivateKey)
		options = append(options, client.WithPrivateKey(key))
	}
	clt, err := client.New(ctx, cfg.GrpcURL, options...)
	if err != nil {
		return nil, err
	}
	return clt, nil
}
