package escrowtasks

import (
	"context"
	"fmt"
	"kwil/internal/pkg/deposits/tasks"
	"kwil/internal/repository"
	"kwil/pkg/chain/contracts/escrow"
)

type task struct {
	contract escrow.EscrowContract
	dao      repository.Queries // this will be used and set for each task
	queries  repository.Queries // this will be set once on initialization

	providerAddress string
}

func New(dao repository.Queries, contract escrow.EscrowContract, providerAddress string) tasks.Runnable {
	return &task{
		contract:        contract,
		dao:             nil,
		queries:         dao,
		providerAddress: providerAddress,
	}
}

func (t *task) Run(ctx context.Context, chunk *tasks.Chunk) error {
	t.dao = t.queries.WithTx(chunk.Tx) // this copies

	err := t.syncDeposits(ctx, chunk)
	if err != nil {
		return fmt.Errorf("error running deposit task: %w", err)
	}

	err = t.syncWithdrawals(ctx, chunk)
	if err != nil {
		return fmt.Errorf("error running withdrawal task: %w", err)
	}

	t.dao = nil // discard the dao

	return nil
}
