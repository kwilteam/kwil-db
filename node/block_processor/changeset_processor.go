package blockprocessor

import (
	"context"
	"fmt"
)

// ChangesetProcessor is a PubSub that listens for changesets and broadcasts them to the receivers.
// Subscribers can be added and removed to listen for changesets.
// Statistics receiver might listen for changesets to update the statistics every block.
// Whereas Network migrations listen for the changesets only during the migration. (that's when you register)
// ABCI --> CS Processor ---> [Subscribers]
// Once all the changesets are processed, all the channels are closed [every block]
// The channels are reset for the next block.
type changesetProcessor struct {
	// channel to receive changesets
	// closed by the pgRepl layer after all the block changes have been pushed to the processor
	csChan chan any

	// subscribers to the changeset processor are the receivers of the changesets
	// Examples: Statistics receiver, Network migration receiver
	subscribers map[string]chan<- any
}

func newChangesetProcessor() *changesetProcessor {
	return &changesetProcessor{
		csChan:      make(chan any, 1), // buffered channel to avoid blocking
		subscribers: make(map[string]chan<- any),
	}
}

// Subscribe adds a subscriber to the changeset processor's subscribers list.
// The receiver can subscribe to the changeset processor using a unique id.
func (c *changesetProcessor) Subscribe(ctx context.Context, id string) (<-chan any, error) {
	_, ok := c.subscribers[id]
	if ok {
		return nil, fmt.Errorf("subscriber with id %s already exists", id)
	}

	ch := make(chan any, 1) // buffered channel to avoid blocking
	c.subscribers[id] = ch
	return ch, nil
}

// Unsubscribe removes the subscriber from the changeset processor.
func (c *changesetProcessor) Unsubscribe(ctx context.Context, id string) error {
	if ch, ok := c.subscribers[id]; ok {
		// close the channel to signal the subscriber to stop listening
		close(ch)
		delete(c.subscribers, id)
		return nil
	}

	return fmt.Errorf("subscriber with id %s does not exist", id)
}

// Broadcast sends changesets to all the subscribers through their channels.
func (c *changesetProcessor) BroadcastChangesets(ctx context.Context) {
	defer c.Close() // All the block changesets have been processed, signal subscribers to stop listening.

	// Listen on the csChan for changesets and broadcast them to all subscribers.
	for cs := range c.csChan {
		for _, ch := range c.subscribers {
			select {
			case ch <- cs:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *changesetProcessor) Close() {
	// c.CsChan is closed by the repl layer (sender closes the channel)
	for _, ch := range c.subscribers {
		close(ch)
	}
}
