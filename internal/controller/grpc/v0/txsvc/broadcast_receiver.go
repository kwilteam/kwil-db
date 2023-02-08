package txsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/kwil/common/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	transactions2 "kwil/pkg/crypto/transactions"
	"kwil/pkg/utils/serialize"
)

// Broadcast handles broadcasted transactions
func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	// convert the transaction
	tx, err := serialize.Convert[commonpb.Tx, transactions2.Transaction](req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	err = tx.Verify() // verify verifies the hash and signature
	if err != nil {
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	// handle the transaction according to its type
	switch tx.PayloadType {
	case transactions2.DEPLOY_DATABASE:
		return s.handleDeployDatabase(ctx, tx)
	case transactions2.MODIFY_DATABASE:
		return nil, fmt.Errorf("not implemented")
	case transactions2.DROP_DATABASE:
		return s.handleDropDatabase(ctx, tx)
	case transactions2.EXECUTE_QUERY:
		return s.handleExecution(ctx, tx)
	case transactions2.WITHDRAW:
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("invalid payload type")
	}
}
