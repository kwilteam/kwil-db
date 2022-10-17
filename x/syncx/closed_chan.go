package syncx

import (
	"context"
	"kwil/x"
)

var closedChanVoid = NewClosedChan[x.Void]()
var closedChanString = NewClosedChan[string]()
var closedChanInt32 = NewClosedChan[int32]()
var closedChanInt64 = NewClosedChan[int64]()

func ClosedChanVoid() Chan[x.Void] {
	return closedChanVoid
}

func ClosedChanString() Chan[string] {
	return closedChanString
}

func ClosedChanInt32() Chan[int32] {
	return closedChanInt32
}

func ClosedChanInt64() Chan[int64] {
	return closedChanInt64
}

func NewClosedChan[T any]() Chan[T] {
	ch_read := make(chan T)
	ch_locked := make(chan x.Void)
	ch_closed := make(chan x.Void)

	close(ch_read)
	close(ch_locked)
	close(ch_closed)

	return &closed_chan[T]{ch_read, ch_locked, ch_closed}
}

type closed_chan[T any] struct {
	ch_read   chan T
	ch_locked chan x.Void
	ch_closed chan x.Void
}

func (c closed_chan[T]) Read() <-chan T {
	return c.ch_read
}

func (closed_chan[T]) Drain(_ context.Context) ([]T, error) {
	return []T{}, nil
}

func (closed_chan[T]) Write(_ T) bool {
	return false
}

func (closed_chan[T]) TryWrite(_ context.Context, _ T) (ok bool, err error) {
	return false, nil
}

func (closed_chan[T]) Close() bool {
	return false
}

func (closed_chan[T]) CloseAndDrain(_ context.Context) ([]T, error) {
	return []T{}, nil
}

func (closed_chan[T]) CloseAndWait(_ context.Context) error {
	return nil
}

func (c closed_chan[T]) ClosedCh() <-chan x.Void {
	return c.ch_closed
}

func (c closed_chan[T]) LockCh() <-chan x.Void {
	return c.ch_locked
}

func (closed_chan[T]) IsClosed() bool {
	return true
}

func (closed_chan[T]) IsLocked() bool {
	return true
}

func (closed_chan[T]) Length() int {
	return 0
}

func (closed_chan[T]) Capacity() int {
	return 0
}
