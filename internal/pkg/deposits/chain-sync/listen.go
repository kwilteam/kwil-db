package chainsync

import (
	"context"
	"go.uber.org/zap"
)

// Listen will listen for new block and update the persistent block number
func (c *chain) listen(ctx context.Context) error {
	blocks := make(chan int64)
	err := c.chainClient.Listen(ctx, blocks)
	if err != nil {
		return err
	}

	go func(c *chain, blocks <-chan int64) {
		c.mu.Lock()
		defer c.mu.Unlock()

		for {
			select {
			case <-ctx.Done():
				c.log.Warn("stopping chain listener", zap.Error(ctx.Err()))
				return
			case block := <-blocks:
				c.log.Debug("new block ", zap.Int64("height", block))
				c.processChunk(ctx, c.height+1, block)
				c.log.Debug("processed chunk", zap.Int64("height", c.height))
			}
		}

	}(c, blocks)

	return nil
}
