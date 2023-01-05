package escrowtasks

import (
	"context"
	"fmt"
	"kwil/x/contracts/escrow"
	"kwil/x/deposits/repository"
	"kwil/x/deposits/service/tasks"
)

type task struct {
	contract escrow.EscrowContract
	dao      *repository.Queries // this will be used and set for each task
	queries  *repository.Queries // this will be set once on initialization
}

func New(dao *repository.Queries, contract escrow.EscrowContract) tasks.Runnable {
	return &task{
		contract: contract,
		dao:      nil,
		queries:  dao,
	}
}

func (t *task) Run(ctx context.Context, chunk *tasks.Chunk) error {
	t.dao = t.queries.WithTx(chunk.Tx) // this copies

	err := t.syncDeposits(ctx, chunk)
	if err != nil {
		return fmt.Errorf("error syncing deposits: %w", err)
	}

	err = t.syncWithdrawals(ctx, chunk)
	if err != nil {
		return fmt.Errorf("error syncing withdrawals: %w", err)
	}

	t.dao = nil // discard the dao

	return nil
}
