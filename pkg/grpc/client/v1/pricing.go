package client

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

func (c *Client) UpdateGasCosts(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	res, err := c.txClient.GasCosts(ctx, &txpb.GasCostsRequest{
		Tx: ConvertTx(tx),
	})
	if err != nil {
		return nil, err
	}

	txRes := ConvertReceipt(res.Receipt)
	return txRes, nil
}
