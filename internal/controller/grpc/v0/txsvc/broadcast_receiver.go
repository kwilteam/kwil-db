package txsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0"
	txpb "kwil/api/protobuf/tx/v0"
	"kwil/pkg/accounts"
	"kwil/pkg/utils/serialize"
)

// Broadcast handles broadcasted transactions
func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	// convert the transaction
	tx, err := serialize.Convert[commonpb.Tx, accounts.Transaction](req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	err = tx.Verify() // verify verifies the hash and signature
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	// handle the transaction according to its type
	switch tx.PayloadType {
	case accounts.DEPLOY_DATABASE:
		return s.handleDeployDatabase(ctx, tx)
	case accounts.MODIFY_DATABASE:
		return nil, fmt.Errorf("not implemented")
	case accounts.DROP_DATABASE:
		return s.handleDropDatabase(ctx, tx)
	case accounts.EXECUTE_QUERY:
		return s.handleExecution(ctx, tx)
	case accounts.WITHDRAW:
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("invalid payload type")
	}
}
