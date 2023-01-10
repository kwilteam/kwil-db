package escrowtasks

import (
	"context"
	"fmt"
	"kwil/x/sqlx/errors"

	"kwil/x/deposits/tasks"
)

func (c *task) syncDeposits(ctx context.Context, chunk *tasks.Chunk) error {
	deposits, err := c.contract.GetDeposits(ctx, chunk.Start, chunk.Finish)
	if err != nil {
		return fmt.Errorf("failed to get deposits from chain: %w", err)
	}

	for _, deposit := range deposits {
		// we first must check if this txid already exists
		// normally we could just check for a unique constraint violation,
		// but since we are doing this in a tx, an error will cancel
		// the whole tx
		id, err := c.dao.GetDepositIdByTx(ctx, deposit.TxHash)
		if err != nil && !errors.IsNoRowsInResult(err) {
			return fmt.Errorf("failed to get deposit by tx: %w", err)

		}

		// if the id is not 0, then there is already a deposit with this txid
		if id != 0 {
			continue
		}

		err = c.dao.Deposit(ctx, deposit)
		if err != nil {
			return fmt.Errorf("failed to deposit: %w", err)
		}
	}

	err = c.dao.CommitDeposits(ctx, chunk.Finish)
	if err != nil {
		return fmt.Errorf("failed to commit deposits: %w", err)
	}

	return nil
}
