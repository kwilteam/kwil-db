package syncx

import (
	"context"
	"kwil/x"
	"sync"
)

type _chan[T any] struct {
	// data_ch used node the primary channel for reading
	// and writing data.
	data_ch chan T

	// lock_ch signals that no more writes are allowed.
	lock_ch chan x.Void

	// closed_ch indicates that the channel is closed
	// and there are no more writers.
	closed_ch chan x.Void

	// mu is used to ensure a consistent for the status
	// and writers count.
	mu sync.RWMutex

	// writers is the number of writers currently
	// writing to the channel.
	writers int32

	// status is the current status of the channel.
	// writable: 0, closed: 1, closed: 2
	status int8
}

// NewChan creates a new Chan.
func NewChan[T any]() Chan[T] {
	return &_chan[T]{
		data_ch:   make(chan T),
		lock_ch:   make(chan x.Void),
		closed_ch: make(chan x.Void),
		mu:        sync.RWMutex{},
	}
}

// NewChanBuffered creates a buffered Chan.
func NewChanBuffered[T any](size int) Chan[T] {
	return &_chan[T]{
		data_ch:   make(chan T, size),
		lock_ch:   make(chan x.Void),
		closed_ch: make(chan x.Void),
		mu:        sync.RWMutex{},
	}
}

func (c *_chan[T]) Read() <-chan T {
	return c.data_ch
}

func (c *_chan[T]) Write(value T) (ok bool) {
	if !c.increment_writers() {
		return false
	}

	defer c.decrement_writers()

	select {
	case c.data_ch <- value:
		return true
	case <-c.lock_ch:
		return false
	}
}

func (c *_chan[T]) TryWrite(ctx context.Context, value T) (ok bool, err error) {
	if !c.increment_writers() {
		return false, nil
	}

	defer c.decrement_writers()

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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.is_closed_unsafe()
}

func (c *_chan[T]) IsLocked() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.is_lockedOrClosed_unsafe()
}

func (c *_chan[T]) Close() bool {
	c.mu.Lock()
	if c.is_lockedOrClosed_unsafe() {
		c.mu.Unlock()
		return false
	}

	if c.writers > 0 {
		c.set_locked_unsafe()
	} else {
		c.set_closed_unsafe()
		defer c.do_close()
	}

	c.mu.Unlock()

	close(c.lock_ch)

	return true
}

func (c *_chan[T]) CloseAndWait(ctx context.Context) error {
	if !c.Close() {
		return nil
	}

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

func (c *_chan[T]) CloseAndDrain(ctx context.Context) ([]T, error) {
	err := c.CloseAndWait(ctx)
	if err != nil {
		return []T{}, err
	}

	return c.Drain(ctx)
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

func (c *_chan[T]) increment_writers() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.is_lockedOrClosed_unsafe() {
		return false
	}

	c.writers++

	return true
}

func (c *_chan[T]) decrement_writers() {
	c.mu.Lock()
	c.writers--
	if c.writers > 0 || !c.is_locked_unsafe() {
		c.mu.Unlock()
		return
	}

	c.mu.TryLock()
	c.set_closed_unsafe()

	c.mu.Unlock()

	c.do_close() // do after unlock
}

func (c *_chan[T]) do_close() {
	close(c.data_ch)
	close(c.closed_ch)
}

func (c *_chan[T]) is_closed_unsafe() bool {
	return c.status == 2
}

func (c *_chan[T]) is_locked_unsafe() bool {
	return c.status == 1
}

func (c *_chan[T]) is_lockedOrClosed_unsafe() bool {
	return c.status > 0
}

func (c *_chan[T]) set_closed_unsafe() {
	c.status = 2
}

func (c *_chan[T]) set_locked_unsafe() {
	c.status = 1
}
