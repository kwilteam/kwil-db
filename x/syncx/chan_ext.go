package syncx

import (
	"context"
	"kwil/x"
	"math"
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
	writers uint32
}

var chan_locked = uint32(2147483648)   // mutex locked bit
var chan_done = uint32(math.MaxUint32) // mutex done

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
	ok, _ = c.TryWrite(nil, value)
	return ok
}

func (c *_chan[T]) TryWrite(ctx context.Context, value T) (ok bool, err error) {
	var writers uint32
	if !c.incrementWriterCount(&writers) {
		return false, nil
	}

	defer c.decrementWriterCount(writers)

	if ctx == nil {
		select {
		case c.data_ch <- value:
			return true, nil
		case <-c.lock_ch:
			return false, nil
		}
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
	return c._isClosed(c.loadWriters())
}

func (c *_chan[T]) IsLocked() bool {
	return c._isLockRequested(c.loadWriters())
}

func (c *_chan[T]) Close() {
	writers := atomic.LoadUint32(&c.writers)
	for {
		if c._isLockRequested(writers) {
			return
		}

		if c._cas(writers, writers|chan_locked) {
			break
		}

		writers = atomic.LoadUint32(&c.writers)
	}

	close(c.lock_ch)
	if writers == 0 {
		c.doClose()
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

func (c *_chan[T]) incrementWriterCount(out *uint32) bool {
	writers := atomic.LoadUint32(&c.writers)
	for {
		if c._isLockRequested(writers) {
			return false
		}

		if c._cas(writers, writers+1) {
			writers++
			break
		}

		writers = atomic.LoadUint32(&c.writers)
	}

	*out = writers

	return true
}

func (c *_chan[T]) decrementWriterCount(writers uint32) {
	if c._cas(writers, writers-1) {
		return
	}

	for {
		writers = atomic.LoadUint32(&c.writers)
		if !c._cas(writers, writers-1) {
			continue
		}

		writers--
		if c._isLockRequested(writers) && !c._hasOutstandingWriters(writers) {
			c.doClose()
			break
		}
	}
}

func (c *_chan[T]) doClose() {
	close(c.data_ch)
	close(c.closed_ch)
}

func (c *_chan[T]) loadWriters() uint32 {
	return atomic.LoadUint32(&c.writers)
}

func (_ *_chan[T]) _isClosed(status uint32) bool {
	return status == math.MaxUint32
}

func (_ *_chan[T]) _hasOutstandingWriters(writers uint32) bool {
	return writers^chan_done > 0
}

func (c *_chan[T]) _isLockRequested(writers uint32) bool {
	return writers&chan_locked == chan_locked
}

func (c *_chan[T]) _cas(expected, updated uint32) bool {
	return atomic.CompareAndSwapUint32(&c.writers, expected, updated)
}
