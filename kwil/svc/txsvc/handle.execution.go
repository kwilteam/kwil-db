package txsvc

import (
	"context"
	"fmt"
	"kwil/x/proto/txpb"
	accountTypes "kwil/x/types/accounts"
	"kwil/x/types/execution"
	"kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

func (s *Service) handleExecution(ctx context.Context, tx *transactions.Transaction) (*txpb.BroadcastResponse, error) {
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

	// get payload
	payload, err := serialize.Deserialize[*execution.ExecutionBody](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type ExecutionBody: %w", err)
	}

	// execute
	err = s.executor.ExecuteQuery(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &txpb.BroadcastResponse{
		Hash: tx.Hash,
		Fee:  price,
	}, nil
}
