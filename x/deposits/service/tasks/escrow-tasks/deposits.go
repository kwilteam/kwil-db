package escrowtasks

import (
	"context"
	"kwil/x/deposits/repository"
	"kwil/x/sqlx/errors"
	"strings"

	"kwil/x/deposits/service/tasks"
)

func (c *task) syncDeposits(ctx context.Context, chunk *tasks.Chunk) error {
	deposits, err := c.contract.GetDeposits(ctx, chunk.Start, chunk.Finish)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		// we first must check if this txid already exists
		// normally we could just check for a unique constraint violation,
		// but since we are doing this in a tx, an error will cancel
		// the whole tx
		id, err := c.dao.GetDepositByTx(ctx, deposit.TxHash)
		if err != nil && !errors.IsNoRowsInResult(err) {
			return err

		}

		// if the id is not 0, then there is already a deposit with this txid
		if id != 0 {
			continue
		}

		err = c.dao.Deposit(ctx, &repository.DepositParams{
			Wallet: strings.ToLower(deposit.Caller),
			Amount: deposit.Amount,
			TxHash: strings.ToLower(deposit.TxHash),
			Height: deposit.Height,
		})
		if err != nil {
			return err
		}
	}

	return c.dao.CommitDeposits(ctx, chunk.Finish)
}
