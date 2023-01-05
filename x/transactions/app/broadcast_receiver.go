package app

import (
	"context"
	"fmt"
	"kwil/x/proto/txpb"
	"kwil/x/transactions"
	"kwil/x/transactions/utils"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	tx := utils.TxFromMsg(req.GetTx())
	err := tx.Verify()
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	// try to execute the transaction
	switch tx.PayloadType {
	case transactions.DEPLOY_DATABASE:
		res, err := s.service.DeployDatabase(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy database: %w", err)
		}

		return &txpb.BroadcastResponse{
			Hash: res.Hash,
			Fee:  res.Fee,
		}, nil
	case transactions.MODIFY_DATABASE:
		return nil, fmt.Errorf("not implemented")
	case transactions.DROP_DATABASE:
		res, err := s.service.DropDatabase(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to drop database: %w", err)
		}

		return &txpb.BroadcastResponse{
			Hash: res.Hash,
			Fee:  res.Fee,
		}, nil
	case transactions.EXECUTE_QUERY:
		res, err := s.service.ExecuteQuery(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}

		return &txpb.BroadcastResponse{
			Hash: res.Hash,
			Fee:  res.Fee,
		}, nil
	case transactions.WITHDRAW:
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("invalid payload type")
	}
}
