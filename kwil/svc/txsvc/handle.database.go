package txsvc

import (
	"context"
	"fmt"
	"kwil/kwil/repository"
	"kwil/x/proto/txpb"
	"kwil/x/types/databases"
	"kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

func (s *Service) handleDeployDatabase(ctx context.Context, tx *transactions.Transaction) (*txpb.BroadcastResponse, error) {
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
	err = s.dao.Spend(ctx, &repository.SpendParams{
		AccountAddress: tx.Sender,
		Balance:        price,
		Nonce:          tx.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spend fee: %w", err)
	}

	db, err := serialize.Deserialize[*databases.Database](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type Database: %w", err)
	}

	err = s.executor.DeployDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy database: %w", err)
	}

	return &txpb.BroadcastResponse{
		Hash: tx.Hash,
		Fee:  tx.Fee,
	}, nil
}

func (s *Service) handleDropDatabase(ctx context.Context, tx *transactions.Transaction) (*txpb.BroadcastResponse, error) {
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
	err = s.dao.Spend(ctx, &repository.SpendParams{
		AccountAddress: tx.Sender,
		Balance:        price,
		Nonce:          tx.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spend fee: %w", err)
	}

	db, err := serialize.Deserialize[*databases.DatabaseIdentifier](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type Database: %w", err)
	}

	err = s.executor.DropDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to drop database: %w", err)
	}

	return &txpb.BroadcastResponse{
		Hash: tx.Hash,
		Fee:  tx.Fee,
	}, nil
}
