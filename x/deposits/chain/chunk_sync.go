package chain

import (
	"context"
	"fmt"
)

func (c *chunk) commit(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := c.dao.SetHeight(ctx, c.finish)
	if err != nil {
		return fmt.Errorf("failed to set height to %d: %w", c.finish, err)
	}

	return c.tx.Commit()
}

// processChunk will process all deposits and withdrawals for a given chunk.
func (c *chain) processChunk(ctx context.Context, start, finish int64) error {
	// the whole chunk should be processed in a single transaction
	tx, err := c.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin db transaction for chunk %d - %d: %w", start, finish, err)
	}
	defer tx.Rollback()

	chunk, err := c.newChunk(ctx, tx, start, finish)
	if err != nil {
		return fmt.Errorf("failed to create chunk: %w", err)
	}

	// process deposits
	err = chunk.syncDeposits(ctx)
	if err != nil {
		return fmt.Errorf("failed to process deposits: %w", err)
	}

	// process withdrawals
	err = chunk.syncWithdrawals(ctx)
	if err != nil {
		return fmt.Errorf("failed to process withdrawals: %w", err)
	}

	// commit the chunk
	err = chunk.commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit chunk: %w", err)
	}

	c.height = finish
	return nil
}
