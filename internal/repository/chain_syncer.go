package repository

import (
	"context"
	"database/sql"
	"fmt"
	depositTypes "kwil/internal/pkg/deposits/types"
	"kwil/internal/repository/gen"
	escrowTypes "kwil/pkg/contracts/escrow/types"
	"strings"
)

type ChainSyncer interface {
	SetHeight(ctx context.Context, chain int32, height int64) error
	GetHeight(ctx context.Context, chain int32) (int64, error)
	GetDepositIdByTx(ctx context.Context, txHash string) (int32, error)
	Deposit(ctx context.Context, deposit *escrowTypes.DepositEvent) error
	CommitDeposits(ctx context.Context, finish int64) error
	ConfirmWithdrawal(ctx context.Context, correlationId string) error
	NewWithdrawal(ctx context.Context, withdrawal *depositTypes.StartWithdrawal) error
	AddTxHashToWithdrawal(ctx context.Context, txHash string, correlationId string) error
}

func (q *queries) SetHeight(ctx context.Context, chain int32, height int64) error {
	return q.gen.SetHeight(ctx, &gen.SetHeightParams{
		Code:   chain,
		Height: height,
	})
}

func (q *queries) GetHeight(ctx context.Context, chain int32) (int64, error) {
	return q.gen.GetHeight(ctx, chain)
}

func (q *queries) GetDepositIdByTx(ctx context.Context, txHash string) (int32, error) {
	return q.gen.GetDepositIdByTx(ctx, txHash)
}

func (q *queries) Deposit(ctx context.Context, deposit *escrowTypes.DepositEvent) error {
	return q.gen.Deposit(ctx, &gen.DepositParams{
		Amount:         deposit.Amount,
		TxHash:         strings.ToLower(deposit.TxHash),
		Height:         deposit.Height,
		AccountAddress: strings.ToLower(deposit.Caller),
	})
}

func (q *queries) CommitDeposits(ctx context.Context, finish int64) error {
	err := q.gen.CommitDeposits(ctx, finish)
	if err != nil {
		return fmt.Errorf("failed to commit deposits: %w", err)
	}

	err = q.gen.DeleteDeposits(ctx, finish)
	if err != nil {
		return fmt.Errorf("failed to delete deposits: %w", err)
	}

	return nil
}

func (q *queries) ConfirmWithdrawal(ctx context.Context, correlationId string) error {
	return q.gen.ConfirmWithdrawal(ctx, correlationId)
}

func (q *queries) NewWithdrawal(ctx context.Context, withdrawal *depositTypes.StartWithdrawal) error {
	return q.gen.NewWithdrawal(ctx, &gen.NewWithdrawalParams{
		CorrelationID:  withdrawal.CorrelationId,
		Amount:         withdrawal.Amount,
		AccountAddress: strings.ToLower(withdrawal.Address),
		Fee:            withdrawal.Fee,
		Expiry:         withdrawal.Expiration,
	})
}

func (q *queries) AddTxHashToWithdrawal(ctx context.Context, txHash string, correlationId string) error {
	return q.gen.AddTxHashToWithdrawal(ctx, &gen.AddTxHashToWithdrawalParams{
		CorrelationID: correlationId,
		TxHash:        sql.NullString{Valid: true, String: strings.ToLower(txHash)},
	})
}

func (q *queries) DeleteDeposits(ctx context.Context, finish int64) error {
	return q.gen.DeleteDeposits(ctx, finish)
}
