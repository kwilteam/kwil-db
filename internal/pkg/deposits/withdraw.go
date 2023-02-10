package deposits

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"kwil/internal/pkg/deposits/types"
	"kwil/pkg/utils/numbers/big"
	"math/rand"
	"time"
)

// StartWithdrawal begins the withdrawal process.  It will alter a user's balance and assign a correlation ID, which will be used to track the withdrawal on-chain.
func (s *depositer) startWithdrawal(ctx context.Context, withdrawal types.WithdrawalRequest) error {
	// start a transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin db transaction for withdrawal: %w", err)
	}
	defer tx.Rollback()

	qtx := s.dao.WithTx(tx)

	// will start by getting the config balance
	account, err := qtx.GetAccount(ctx, withdrawal.Address)
	if err != nil {
		return err
	}

	// compare the balance to the withdrawal amount
	cmp, err := big.BigStr(withdrawal.Amount).Compare(withdrawal.Amount)
	if err != nil {
		return err
	}

	// if the balance is less than the withdrawal amount, set the withdrawal amount to the balance
	if cmp < 0 {
		withdrawal.Amount = account.Balance
	}

	// get the current block height
	blockHeight, err := qtx.GetHeight(ctx, int32(s.chain.ChainCode()))
	if err != nil {
		return err
	}

	// generate a correlation id
	correlationId, err := generateCid(16, withdrawal.Address)
	if err != nil {
		return err
	}

	err = qtx.NewWithdrawal(ctx, &types.StartWithdrawal{
		CorrelationId: correlationId,
		Address:       account.Address,
		Amount:        withdrawal.Amount,
		Fee:           account.Spent,
		Expiration:    blockHeight + s.expirationPeriod,
	})

	if err != nil {
		return err
	}

	return tx.Commit()
}

// confirmWithdrawal confirms a withdrawal.  This is called after the withdrawal has been mined and finalized on the blockchain.
// It identifies the withdrawal by the correlation ID and marks it as confirmed.
func (s *depositer) confirmWithdrawal(ctx context.Context, correlationId string) error {
	return s.dao.ConfirmWithdrawal(ctx, correlationId)
}

/*
// expireWithdrawals expires withdrawals that have an expiry block height less than or equal to the given height.
func (s *depositsService) expireWithdrawals(ctx context.Context, height int64) error {
	return s.dao.ExpireWithdrawals(ctx, height)
}
*/

// generateCid generates a correlation id for the withdrawal.
// it takes a desired length as well as a string to seed the random number generator
func generateCid(l uint8, str string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		return "", fmt.Errorf("failed to write to hash for seeding correlation_id: %w", err)
	}

	seed := binary.LittleEndian.Uint64(h.Sum(nil))
	rand.Seed(int64(seed))
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, l)
	for i := uint8(0); i < l; i++ {
		result[i] = types.CorrelationIdCharacters[rand.Intn(len(types.CorrelationIdCharacters))]
	}
	return string(result), nil
}
