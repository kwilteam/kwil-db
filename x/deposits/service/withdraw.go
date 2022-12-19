package service

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
	"math/big"
	"math/rand"
	"time"
)

// StartWithdrawal begins the withdrawal process.  It will alter a user's balance and assign a correlation ID, which will be used to track the withdrawal on-chain.
func (s *depositsService) StartWithdrawal(ctx context.Context, withdrawal dto.StartWithdrawal) error {
	// start a transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin db transaction for withdrawal: %w", err)
	}
	defer tx.Rollback()

	qtx := s.doa.WithTx(tx)

	// will start by getting the wallet balance
	wallet, err := qtx.GetBalanceAndSpent(ctx, withdrawal.Wallet)
	if err != nil {
		return err
	}

	// compare the balance to the withdrawal amount
	cmp, err := compareBigIntStrings(wallet.Balance, withdrawal.Amount)
	if err != nil {
		return err
	}

	// if the balance is less than the withdrawal amount, set the withdrawal amount to the balance
	if cmp < 0 {
		withdrawal.Amount = wallet.Balance
	}

	// get the current block height
	blockHeight, err := qtx.GetHeight(ctx)
	if err != nil {
		return err
	}

	// generate a correlation id
	correlationId, err := generateCid(16, withdrawal.Wallet)
	if err != nil {
		return err
	}

	err = qtx.NewWithdrawal(ctx, &repository.NewWithdrawalParams{
		CorrelationID: correlationId,
		WalletID:      wallet.ID,
		Amount:        withdrawal.Amount,
		Fee:           wallet.Spent,
		Expiry:        blockHeight + s.expirationPeriod,
	})

	if err != nil {
		return err
	}

	return tx.Commit()
}

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
		result[i] = dto.CidCharacters[rand.Intn(len(dto.CidCharacters))]
	}
	return string(result), nil
}

func compareBigIntStrings(a, b string) (int, error) {
	// convert to big.Int
	ai, ok := new(big.Int).SetString(a, 10)
	if !ok {
		return 0, fmt.Errorf("failed to convert %s to big.Int", a)
	}
	bi, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return 0, fmt.Errorf("failed to convert %s to big.Int", b)
	}

	// compare
	return ai.Cmp(bi), nil
}
