package chainsync

import (
	"kwil/internal/pkg/deposits/tasks"
)

// used for adding new tasks (e.g. listening to events from other smart contracts)
func (c *chain) RegisterTask(task tasks.Runnable) {
	c.tasks.Add(task)
}
