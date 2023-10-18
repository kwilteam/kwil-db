package client

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"google.golang.org/grpc/status"
)

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error) {
	pbTx := convertTx(tx)
	res, err := c.txClient.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		statusError, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("unrecognized broadcast failure: %w", err)
		}

		code, message := statusError.Code(), statusError.Message()
		err = fmt.Errorf("%v (%d)", message, code)

		for _, detail := range statusError.Details() {
			if bcastErr, ok := detail.(*txpb.BroadcastErrorDetails); ok {
				switch txCode := transactions.TxCode(bcastErr.Code); txCode {
				case transactions.CodeWrongChain:
					err = errors.Join(err, transactions.ErrWrongChain)
				case transactions.CodeInvalidNonce:
					err = errors.Join(err, transactions.ErrInvalidNonce)
				default:
					err = errors.Join(err, errors.New(txCode.String()))
				}
			} else { // else unknown details type
				err = errors.Join(err, fmt.Errorf("unrecognized status error detail type %T", detail))
			}
		}
		return nil, err
	}

	return res.GetTxHash(), nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	res, err := c.txClient.Ping(ctx, &txpb.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}

func (c *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	// convert transaction to proto
	pbTx := convertTx(tx)

	res, err := c.txClient.EstimatePrice(ctx, &txpb.EstimatePriceRequest{
		Tx: pbTx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate cost: %w", err)
	}

	bigCost, ok := new(big.Int).SetString(res.Price, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert price to big.Int")
	}

	return bigCost, nil
}
