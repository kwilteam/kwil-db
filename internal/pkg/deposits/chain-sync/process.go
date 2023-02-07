package chainsync

import (
	"context"
	"fmt"
	"kwil/internal/pkg/deposits/tasks"
)

// processChunk will run all tasks for a chunk.
func (c *chain) processChunk(ctx context.Context, start, finish int64) error {
	// the whole chunk should be processed in a single transaction
	tx, err := c.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin db transaction for chunk %d - %d: %w", start, finish, err)
	}
	defer tx.Rollback()

	chunk := &tasks.Chunk{
		Start:  start,
		Finish: finish,
		Tx:     tx,
	}

	err = c.tasks.Run(ctx, chunk)
	if err != nil {
		return fmt.Errorf("failed to run tasks for chunk %d - %d: %w", start, finish, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit chunk %d - %d: %w", start, finish, err)
	}

	c.height = finish
	return nil
}
