package escrowtasks

import (
	"context"
	"fmt"
	"kwil/x/deposits/tasks"
	"kwil/x/types/accounts"
	bigutil "kwil/x/utils/big"
	"math/big"
)

// syncWithdrawals will sync all withdrawals for the chunk to the database
func (c *task) syncWithdrawals(ctx context.Context, chunk *tasks.Chunk) error {

	withdrawals, err := c.contract.GetWithdrawals(ctx, chunk.Start, chunk.Finish)
	if err != nil {
		return fmt.Errorf("failed to get withdrawals from chain: %w", err)
	}
	// for the withdrawals, we will simply decrease the balance by the total amount.
	for _, withdrawal := range withdrawals {
		// first try to confirm the withdrawal.
		// if the correlationID is not found in the withdrawal table, then
		// the balance for this withdrawal still exists in the wallet.
		// If the balance still exists in the wallet, then we will start by zeroing
		// the wallet's spent, and once that is 0, will decrease the balance.
		err = c.dao.ConfirmWithdrawal(ctx, withdrawal.Cid)
		// if there is no error, then the withdrawal has simply been confirmed and we can move on.
		// if there is an error, then the withdrawal had not been started on this machine,
		// so we will subtract the whole amount.
		if err == nil {
			continue
		}

		// we will subtract the fee from spent and the amount from balance.
		// if the spent reaches 0, then the rest of the fees will be subtracted from the balance.
		wallet, err := c.dao.GetAccount(ctx, withdrawal.Caller)
		if err != nil {
			return fmt.Errorf("failed to get balance and spent for wallet %s: %w", withdrawal.Caller, err)
		}

		// calculate the new balance and spent.
		newBalance, newSpent, err := calculateNewBalances(wallet.Balance, wallet.Spent, withdrawal.Amount, withdrawal.Fee)
		if err != nil {
			return fmt.Errorf("failed to calculate new balance and spent for wallet %s: %w", withdrawal.Caller, err)
		}

		// now that we have the new balance and spent, we can update the wallet.
		err = c.dao.UpdateAccount(ctx, &accounts.Account{
			Address: wallet.Address,
			Balance: newBalance.String(),
			Spent:   newSpent.String(),
		})
		if err != nil {
			return fmt.Errorf("failed to set balance and spent for wallet %s: %w", withdrawal.Caller, err)
		}
	}
	return nil
}

// calculatNewBalances will calculate what a wallet's new balance and spent will be after a withdrawal.
func calculateNewBalances(balance, spent, amount, fee string) (*big.Int, *big.Int, error) {
	// let's try to subtract the fee from the spent first.
	newSpent, err := bigutil.BigStr(spent).Sub(fee)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subtract fee from spent: %w", err)
	}

	newBalance, err := bigutil.BigStr(balance).Sub(amount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subtract amount from balance: %w", err)
	}

	if newSpent.Sign() < 0 {
		// if spent is less than 0, any remaining fees will be subtracted from the balance.
		newBalance = newBalance.Sub(newBalance, newSpent.Abs(newSpent))
	}

	// lastly, check if the balance is less than 0.
	// this should never happen; if it does, then there is a bug in the code.
	if newBalance.Sign() < 0 {
		return nil, nil, fmt.Errorf("balance is less than 0: %s", newBalance.String())
	}

	return newBalance, newSpent, err
}
