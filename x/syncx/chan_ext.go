package syncx

import (
	"context"
	"kwil/x"
	"math"
	"sync"
	"sync/atomic"
)

type _chan[T any] struct {
	// data_ch used as the primary channel for reading
	// and writing data.
	data_ch chan T

	// lock_ch signals that no more writes should be allowed.
	lock_ch chan x.Void

	// closed_ch indicates that the channel is closed
	// and there are no more writers.
	closed_ch chan x.Void

	// writer_and_lock_state keeps track of the # of
	// concurrent writers, whether a close request has
	// been submitted, or whether the channel is closed.
	writer_and_lock_state int32

	mu sync.Mutex
}

// NewChan creates a new Chan.
func NewChan[T any]() Chan[T] {
	return &_chan[T]{
		data_ch:   make(chan T),
		lock_ch:   make(chan x.Void),
		closed_ch: make(chan x.Void),
	}
}

// NewChanBuffered creates a buffered Chan.
func NewChanBuffered[T any](size int) Chan[T] {
	return &_chan[T]{
		data_ch:   make(chan T, size),
		lock_ch:   make(chan x.Void),
		closed_ch: make(chan x.Void),
	}
}

func (c *_chan[T]) Read() <-chan T {
	return c.data_ch
}

func (c *_chan[T]) Drain(ctx context.Context) ([]T, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	done := false
	var values []T
	for !done {
		select {
		case value, ok := <-c.data_ch:
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
	case c.data_ch <- value:
		return true
	case <-c.lock_ch:
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
	case c.data_ch <- value:
		return true, nil
	case <-c.lock_ch:
		return false, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (c *_chan[T]) IsClosed() bool {
	return c._isClosed(c.loadCnt())
}

func (c *_chan[T]) IsLocked() bool {
	return c._isLockRequested(c.loadCnt())
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
		<-c.closed_ch
		return nil
	}

	select {
	case <-c.closed_ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *_chan[T]) ClosedCh() <-chan x.Void {
	return c.closed_ch
}

func (c *_chan[T]) LockCh() <-chan x.Void {
	return c.lock_ch
}

func (c *_chan[T]) Length() int {
	return len(c.data_ch)
}

func (c *_chan[T]) Capacity() int {
	return cap(c.data_ch)
}

func (c *_chan[T]) acquireWriteLatch(out *int32) bool {
	for {
		latchCnt := atomic.LoadInt32(&c.writer_and_lock_state)
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

		// No writers, we attempt to update the writer_and_lock_state to closed
		if !c._cas(0, math.MinInt32) {
			continue
		}

		c.doClose()
		return
	}
}

func (c *_chan[T]) setToTerminalState(latchCnt int32) bool {
	if c._hasOutstandingWriters(latchCnt) {
		// Attempt to set to a close requested state
		if !c._cas(latchCnt, latchCnt*-1) {
			return false
		}

		// writers present, signal them to exit
		close(c.lock_ch)

		return true
	}

	// No writers, attempting to update the writer_and_lock_state to closed
	if !c._cas(0, math.MinInt32) {
		return false
	}

	close(c.lock_ch) // no writers, but clean-up anyway
	c.doClose()      // Close was successful, go ahead and do a full close

	return true
}

func (c *_chan[T]) doClose() {
	close(c.data_ch)
	close(c.closed_ch)
}

func (c *_chan[T]) loadCnt() int32 {
	return atomic.LoadInt32(&c.writer_and_lock_state)
}

func (_ *_chan[T]) _isClosed(status int32) bool {
	return status == math.MinInt32
}

func (_ *_chan[T]) _isTerminalState(status int32) bool {
	return status < 0
}

func (_ *_chan[T]) _hasOutstandingWriters(status int32) bool {
	return status != 0
}

func (c *_chan[T]) _isLockRequested(status int32) bool {
	return status < 0 && !c._isClosed(status)
}

func (c *_chan[T]) _cas(expected, updated int32) bool {
	return atomic.CompareAndSwapInt32(&c.writer_and_lock_state, expected, updated)
}
