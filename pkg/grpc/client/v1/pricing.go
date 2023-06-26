package client

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) UpdateGasCosts(ctx context.Context, gas_enabled bool) error {
	_, err := c.txClient.GasCosts(ctx, &txpb.GasCostsRequest{Enabled: gas_enabled})
	return err
}
