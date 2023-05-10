package client

import (
	"context"
	"fmt"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) GetConfig(ctx context.Context) (*SvcConfig, error) {
	res, err := c.txClient.GetConfig(ctx, &txpb.GetConfigRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return &SvcConfig{
		ChainCode:       res.ChainCode,
		PoolAddress:     res.PoolAddress,
		ProviderAddress: res.ProviderAddress,
	}, nil
}

type SvcConfig struct {
	ChainCode       int64
	PoolAddress     string
	ProviderAddress string
}
