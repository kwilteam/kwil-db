package syncx

import (
	"context"
	"kwil/x/rx"
	"math"
	"sync/atomic"
)

type _chan[T any] struct {
	// ch used as the primary channel for reading
	// and writing data.
	ch chan T

	// quit signals the closure sequence has started
	// for the channel.
	quit chan rx.Void

	// done indicates that the channel is closed
	done chan rx.Void

	// latchCnt keeps track of the # of concurrent
	// writers, whether a close request has been
	// submitted, or whether the channel is closed.
	latchCnt int32
}

// NewChan creates a new Chan.
func NewChan[T any]() Chan[T] {
	return &_chan[T]{
		ch:   make(chan T),
		quit: make(chan rx.Void),
		done: make(chan rx.Void),
	}
}

// NewChanBuffered creates a buffered Chan.
func NewChanBuffered[T any](size int) Chan[T] {
	return &_chan[T]{
		ch:   make(chan T, size),
		quit: make(chan rx.Void),
		done: make(chan rx.Void),
	}
}

func (c *_chan[T]) Read() <-chan T {
	return c.ch
}

func (c *_chan[T]) Drain(ctx context.Context) ([]T, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	done := false
	var values []T
	for !done {
		select {
		case value, ok := <-c.ch:
			if !ok {
				done = true
			} else {
				values = append(values, value)
			}
		case <-ctx.Done():
			return []T{}, ctx.Err()
		default:
			done = true
		}
	}

	if values != nil {
		return values, nil
	}

	return []T{}, nil
}

func (c *_chan[T]) Write(value T) (ok bool) {
	var latchCnt int32
	if !c.acquireWriteLatch(&latchCnt) {
		return false
	}

	defer c.releaseWriteLatch(latchCnt)

	select {
	case c.ch <- value:
		return true
	case <-c.quit:
		return false
	}
}

func (c *_chan[T]) TryWrite(ctx context.Context, value T) (ok bool, err error) {
	var latchCnt int32
	if !c.acquireWriteLatch(&latchCnt) {
		return false, nil
	}

	defer c.releaseWriteLatch(latchCnt)

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case c.ch <- value:
		return true, nil
	case <-c.quit:
		return false, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (c *_chan[T]) IsDone() bool {
	return c._isClosed(c.loadCnt())
}

func (c *_chan[T]) IsClosing() bool {
	return c._isCloseRequested(c.loadCnt())
}

func (c *_chan[T]) Close() {
	for {
		latchCnt := c.loadCnt()
		if c._isTerminalState(latchCnt) {
			return
		}

		if c.setToTerminalState(latchCnt) {
			return
		}
	}
}

func (c *_chan[T]) CloseAndWait(ctx context.Context) error {
	c.Close()

	if ctx == nil {
		<-c.done
		return nil
	}

	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *_chan[T]) ClosedCh() <-chan rx.Void {
	return c.done
}

func (c *_chan[T]) Length() int {
	return len(c.ch)
}

func (c *_chan[T]) Capacity() int {
	return cap(c.ch)
}

func (c *_chan[T]) acquireWriteLatch(out *int32) bool {
	for {
		latchCnt := atomic.LoadInt32(&c.latchCnt)
		if c._isTerminalState(latchCnt) {
			return false
		}

		*out = latchCnt + 1
		if c._cas(latchCnt, *out) {
			return true
		}
	}
}

func (c *_chan[T]) releaseWriteLatch(latchCnt int32) {
	if c._cas(latchCnt, latchCnt-1) {
		return
	}

	for {
		latchCnt = c.loadCnt()
		if latchCnt > 0 {
			if c._cas(latchCnt, latchCnt-1) {
				return
			}
			continue
		}

		next := latchCnt + 1
		if next != 0 {
			if c._cas(latchCnt, next) {
				return
			}
			continue
		}

		// No writers, so we attempt to update the latchCnt to closed
		if !c._cas(0, math.MinInt32) {
			continue
		}

		c.doClose()
		return
	}
}

func (c *_chan[T]) setToTerminalState(latchCnt int32) bool {
	if c._hasOutstandingLatches(latchCnt) {
		// Attempt to set to a close requested state
		if !c._cas(latchCnt, latchCnt*-1) {
			return false
		}

		// writers present, so signal them to exit
		close(c.quit)

		return true
	}

	// No writers, so we attempt to update the latchCnt to closed
	if !c._cas(0, math.MinInt32) {
		return false
	}

	close(c.quit) // no writers, but clean-up anyway
	c.doClose()   // Close was successful, so we go ahead and do a full close

	return true
}

func (c *_chan[T]) doClose() {
	close(c.ch)
	close(c.done)
}

func (c *_chan[T]) loadCnt() int32 {
	return atomic.LoadInt32(&c.latchCnt)
}

func (_ *_chan[T]) _isClosed(status int32) bool {
	return status == math.MinInt32
}

func (_ *_chan[T]) _isTerminalState(status int32) bool {
	return status < 0
}

func (_ *_chan[T]) _hasOutstandingLatches(status int32) bool {
	return status != 0
}

func (c *_chan[T]) _isCloseRequested(status int32) bool {
	return status < 0 && !c._isClosed(status)
}

func (c *_chan[T]) _cas(expected, updated int32) bool {
	return atomic.CompareAndSwapInt32(&c.latchCnt, expected, updated)
}
