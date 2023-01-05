package service

import (
	"context"
	"fmt"
	accountDto "kwil/x/accounts/dto"
	execDto "kwil/x/execution/dto"
	"kwil/x/transactions/dto"
	"kwil/x/transactions/utils"
)

func (s *service) DeployDatabase(ctx context.Context, tx *dto.Transaction) (*dto.Response, error) {
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

	db, err := utils.DecodePayload[*execDto.Database](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type Database: %w", err)
	}

	err = s.execution.DeployDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy database: %w", err)
	}

	return &dto.Response{
		Hash: tx.Hash,
		Fee:  price,
	}, nil
}

func (s *service) DropDatabase(ctx context.Context, tx *dto.Transaction) (*dto.Response, error) {
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

	db, err := utils.DecodePayload[*execDto.DatabaseIdentifier](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type DatabaseIdentifier: %w", err)
	}

	err = s.execution.DropDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to drop database: %w", err)
	}

	return &dto.Response{
		Hash: tx.Hash,
		Fee:  price,
	}, nil
}
