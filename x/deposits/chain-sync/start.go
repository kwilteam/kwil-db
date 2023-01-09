package chainsync

import "context"

func (c *chain) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.sync(ctx)
	if err != nil {
		return err
	}

	return c.listen(ctx)
}
