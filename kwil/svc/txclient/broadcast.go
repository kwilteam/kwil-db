package txclient

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0/gen/go"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

func (c *client) Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.Response, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[transactions.Transaction, commonpb.Tx](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.txs.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to broadcast transaction: %w", err)
	}

	// convert response to transaction
	txRes, err := serialize.Convert[txpb.BroadcastResponse, transactions.Response](res)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	return txRes, nil
}
