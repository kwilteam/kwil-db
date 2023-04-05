package txsvc

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
	kTx "kwil/pkg/tx"
	"kwil/pkg/utils/serialize"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	tx, err := serialize.Convert[txpb.Tx, kTx.Transaction](req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	err = tx.Verify()
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	switch tx.PayloadType {
	case kTx.DEPLOY_DATABASE:
		return handleReceipt(s.deploy(ctx, tx))
	case kTx.DROP_DATABASE:
		return handleReceipt(s.drop(ctx, tx))
	case kTx.EXECUTE_ACTION:
		return handleReceipt(s.executeAction(ctx, tx))
	default:
		return nil, fmt.Errorf("invalid payload type")
	}
}

func handleReceipt(r *kTx.Receipt, err error) (*txpb.BroadcastResponse, error) {
	if err != nil {
		return nil, err
	}

	return &txpb.BroadcastResponse{
		Receipt: &txpb.TxReceipt{
			TxHash: r.TxHash,
			Fee:    r.Fee,
			Body:   r.Body,
		},
	}, nil
}
