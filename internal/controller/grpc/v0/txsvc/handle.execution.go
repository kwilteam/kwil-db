package txsvc

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v0"
	accountTypes "kwil/pkg/accounts"
	"kwil/pkg/databases/clean"
	"kwil/pkg/databases/executables"
	"kwil/pkg/utils/serialize"
)

func (s *Service) handleExecution(ctx context.Context, tx *accountTypes.Transaction) (*txpb.BroadcastResponse, error) {
	// get the fee
	price, err := s.pricing.GetPrice(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	ok, err := checkFee(tx.Fee, price)
	if err != nil {
		return nil, fmt.Errorf("failed to check fee: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("fee is not enough")
	}

	// try to spend the fee
	err = s.dao.Spend(ctx, &accountTypes.Spend{
		Address: tx.Sender,
		Amount:  price,
		Nonce:   tx.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spend fee: %w", err)
	}

	// get executionBody
	executionBody, err := serialize.Deserialize[*executables.ExecutionBody](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type ExecutionBody: %w", err)
	}

	clean.Clean(&executionBody)

	// execute
	err = s.executor.ExecuteQuery(ctx, executionBody, tx.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &txpb.BroadcastResponse{
		Hash: tx.Hash,
		Fee:  price,
	}, nil
}
