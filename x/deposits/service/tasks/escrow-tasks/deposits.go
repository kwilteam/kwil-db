package escrowtasks

import (
	"context"
	"kwil/x/deposits/repository"

	"kwil/x/deposits/service/tasks"
)

func (c *task) syncDeposits(ctx context.Context, chunk *tasks.Chunk) error {
	deposits, err := c.contract.GetDeposits(ctx, chunk.Start, chunk.Finish)
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

	return c.dao.CommitDeposits(ctx, chunk.Finish)
}
