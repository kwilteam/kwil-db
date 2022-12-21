package chain

import (
	"context"
	"fmt"
)

// Sync synchronizes the deposits and withdrawals up to the latest block.
func (c *chain) Sync(ctx context.Context) error {
	// get the last synced height
	lastSynced, err := c.dao.GetHeight(ctx)
	if err != nil {
		return err
	}

	// get the latest confirmed block
	latest, err := c.chainClient.GetLatestBlock(ctx, true)
	if err != nil {
		return err
	}

	// split into chunks of n blocks
	chunks := splitBlocks(lastSynced, latest, c.chunkSize)

	for _, chnk := range chunks {
		err = c.processChunk(ctx, chnk[0], chnk[1])
		if err != nil {
			return fmt.Errorf("failed to process chunk: %w", err)
		}
	}
	return nil
}

/*
split into chunks of n blocks

e.g. if we are at block 0 and the last block is 350,000 and chunkRange-size is 100,000,
we will process [0, 99999] [100000, 199999], [200000, 299999], [300000, 350000]

the last chunkRange should have an additional block added to it
*/
type chunkRange [2]int64

func splitBlocks(start, end, chunkSize int64) []chunkRange {
	if start == end {
		return []chunkRange{{start, start}}
	}
	var chunks []chunkRange
	for i := start; i < end; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > end {
			chunkEnd = end
		}
		chunks = append(chunks, chunkRange{i, chunkEnd - 1})
	}

	if chunks[len(chunks)-1][1] != end {
		chunks[len(chunks)-1][1] = end
	}
	return chunks
}
