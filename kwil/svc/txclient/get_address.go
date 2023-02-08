package txclient

import (
	"context"
	"kwil/x/proto/txpb"
)

func (c *client) GetValidatorAddress(ctx context.Context) (string, error) {
	res, err := c.txs.GetAddress(ctx, &txpb.GetAddressRequest{})
	if err != nil {
		return "", err
	}

	return res.Address, nil
}
