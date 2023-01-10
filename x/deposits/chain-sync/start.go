package chainsync

import (
	"context"
	"fmt"
)

func (c *chain) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.sync(ctx)
	if err != nil {
		return fmt.Errorf("error syncing chain: %w", err)
	}

	return c.listen(ctx)
}
