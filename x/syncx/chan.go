package syncx

import (
	"context"
	"kwil/x/rx"
)

// Chan a safe channel abstraction that allows for closure
// without a panic on close.
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
	// If the context is nil, it will be ignored (i.e. it
	// will behave the same as Write).
	TryWrite(ctx context.Context, value T) (ok bool, err error)

	// Close will either close the channel or request
	// closure if there are active writers.
	Close()

	// CloseAndWait will either close the channel or request
	// closure if there are active writers. The call will
	// BLOCK until all writers have completed or the context
	// is cancelled. If the context is nil, the call will
	// only block until all writers have completed.
	CloseAndWait(ctx context.Context) error

	// ClosedCh returns a channel to use when needing to
	// await Chan closure. This channel will be closed,
	// but may still have items in the channel that can be
	// read/drained.
	ClosedCh() <-chan rx.Void

	// IsDone returns true if the channel no longer writable
	// or readable
	IsDone() bool

	// IsClosing returns true if the channel is no longer
	// writeable but can still be read from.
	IsClosing() bool

	// Length returns the number of items in the channel.
	// If the channel is closed, the length returned will
	// be equal to the items still remaining if buffered.
	Length() int

	// Capacity returns zero if the channel is not buffered,
	// else it will return the capacity of the channel
	// specified when created (regardless if it is closed).
	Capacity() int
}
