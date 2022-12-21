package chain

import "context"

// Listen will listen for new block and update the persistent block number
func (c *chain) Listen(ctx context.Context) error {
	blockChan, err := c.chainClient.Listen(ctx, true)
	if err != nil {
		return err
	}

	go func(c *chain, blockChan <-chan int64) {
		c.mu.Lock()
		defer c.mu.Unlock()

		for {
			select {
			case <-ctx.Done():
				return
			case block := <-blockChan:
				c.processChunk(ctx, c.height+1, block)
			}
		}

	}(c, blockChan)

	return nil
}
