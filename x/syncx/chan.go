package syncx

import (
	"context"
	"kwil/x"
)

// Chan a channel abstraction that supports:
//  1. Closure without panic on duplicate close.
//  2. Write checks without panic on a closed channel.
//  3. Events for write locking and channel closure.
//  4. Convenience methods for draining, close w/ block,
//     close w/ timeout, and capacity/length.
type Chan[T any] interface {

	// Read returns a channel to use for reading items
	// emitted via Write.
	Read() <-chan T

	// Drain reads all items from the channel and returns
	// them as a slice.
	Drain(ctx context.Context) ([]T, error)

	// Write will put a value in the channel and return
	// true if the channel is still open and not in
	// a close requested state, else it will return false.
	Write(value T) bool

	// TryWrite will put a value in the channel and return
	// true if the channel is still open and not in
	// a close requested state, else it will return false.
	// If the context is nil, the method will effectively
	// behave the same as Write.
	TryWrite(ctx context.Context, value T) (ok bool, err error)

	// Close will either close the channel or lock
	// if there are active writers.
	Close()

	// CloseAndWait will either close the channel or lock
	// if there are active writers. The call will BLOCK
	// until all writers have completed or the context
	// is cancelled. If the context is nil, the call will
	// only block until all writers have completed.
	CloseAndWait(ctx context.Context) error

	// ClosedCh returns a channel to use when needing to
	// await Chan closure. This channel will be closed,
	// but may still have items in the channel that can be
	// read/drained.
	ClosedCh() <-chan x.Void

	// LockCh returns a channel to use when needing to
	// respond to the Chan being locked for writes. This
	// allows a writer to halt concurrent writes that may
	// be competing/contentious with the channel closure.
	// Any items in the channel can still be read/drained.
	LockCh() <-chan x.Void

	// IsClosed returns true if the channel no longer writable
	// and no longer has any in-flight writers. Items can
	// still be read/drained from the channel.
	IsClosed() bool

	// IsLocked returns true if the channel currently has
	// in-flight writers but is no longer writeable. Items
	// can still be read/drained from the channel.
	IsLocked() bool

	// Length returns the number of items in the channel.
	// If the channel is closed, the length returned will
	// be equal to the items still remaining if buffered.
	Length() int

	// Capacity returns zero if the channel is not buffered,
	// else it will return the capacity of the channel
	// specified when created (regardless if it is closed).
	Capacity() int
}
