package service

import (
	"context"
	"fmt"
	accountDto "kwil/x/accounts/dto"
	execDto "kwil/x/execution/dto"
	"kwil/x/transactions/dto"
	"kwil/x/utils/serialize"
)

func (s *service) ExecuteQuery(ctx context.Context, tx *dto.Transaction) (*dto.Response, error) {
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
	err = s.accounts.Spend(ctx, &accountDto.Spend{
		Amount:  price,
		Address: tx.Sender,
		Nonce:   tx.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spend fee: %w", err)
	}

	payload, err := serialize.Deserialize[*execDto.ExecutionBody](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type ExecutionBody: %w", err)
	}

	err = s.execution.ExecuteQuery(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &dto.Response{
		Hash: tx.Hash,
		Fee:  price,
	}, nil
}
