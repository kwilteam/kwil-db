package pub

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
)

type message_with_ctx struct {
	ctx     context.Context
	msg     *kgo.Record
	ackNack AckNackFn
}

func (c *message_with_ctx) fail(err error) {
	a := c.ackNack
	if a != nil {
		a(err)
	}
}

func (c *message_with_ctx) completeOrFail(err error) bool {
	a := c.ackNack
	if a == nil {
		return false
	}

	a(err)

	return true
}
