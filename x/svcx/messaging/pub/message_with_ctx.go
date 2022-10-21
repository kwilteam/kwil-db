package pub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"sync/atomic"
	"unsafe"
)

type message_with_ctx struct {
	ctx     context.Context
	msg     *kgo.Record
	ackNack *AckNackFn
}

func (c *message_with_ctx) fail(err error) {
	c.completeOrFail(err)
}

func (c *message_with_ctx) completeOrFail(err error) (out bool) {
	a := c.ackNack
	if a == nil || a == &none_ack {
		return false
	}

	ptr := unsafe.Pointer(c.ackNack)
	if !atomic.CompareAndSwapPointer(&ptr, unsafe.Pointer(a), unsafe.Pointer(&none_ack)) {
		return false
	}

	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				e = fmt.Errorf("unknown panic: %v", r)
			}
			fmt.Println("panic in ack/nack:", e)
			out = false
		}
	}()

	fn := *a
	fn(err)

	return true
}
