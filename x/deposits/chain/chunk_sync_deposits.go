package chain

import (
	"context"
	"kwil/x/deposits/repository"
)

// syncDeposits will sync all deposits from the chain, to the deposits table, and then commit to the wallets table.
// this all occurs within the chunk's transaction, and should be rolled back if any errors occur in future steps.
func (c *chunk) syncDeposits(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	deposits, err := c.chainClient.GetDeposits(ctx, c.start, c.finish)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		err := c.dao.Deposit(ctx, &repository.DepositParams{
			Wallet: deposit.Caller,
			Amount: deposit.Amount,
			TxHash: deposit.TxHash,
			Height: deposit.Height,
		})
		if err != nil {
			return err
		}
	}

	return c.dao.CommitDeposits(ctx, c.finish)
}
