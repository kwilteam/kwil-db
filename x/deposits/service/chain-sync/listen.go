package chainsync

import "context"

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
				return
			case block := <-blocks:
				c.processChunk(ctx, c.height+1, block)
			}
		}

	}(c, blocks)

	return nil
}
