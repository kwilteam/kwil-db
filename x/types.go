package x

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

var _closedChan chan Void

func init() {
	ch := make(chan Void)
	close(ch)
	_closedChan = ch
}

// Void An empty struct{} typed as 'Void'
type Void struct{}

// ClosedChanVoid is a closed Void channel that can be
// used in default closed use cases or in context
// of atomic CAS operations for closure scenarios.
func ClosedChanVoid() chan Void {
	return _closedChan
}

// Tuple2 is a combination of two values of any type.
// In general, it is used in place of declaring one-off
// structs for passing around a pair of values.
type Tuple2[T, U any] struct {
	P1 T
	P2 U
}

// Tuple3 is a combination of three values of any type.
// In general, it is used in place of declaring one-off
// structs for passing around three values.
type Tuple3[T, U, V any] struct {
	P1 T
	P2 U
	P3 V
}

// Tuple4 is a combination of four values of any type.
// In general, it is used in place of declaring one-off
// structs for passing around four values.
type Tuple4[T, U, V, W any] struct {
	P1 T
	P2 U
	P3 V
	P4 W
}

// Integer is a type that represents the various
// uint and int golang types
type Integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Closeable is an interface that represents a type
// that can be closed
type Closeable interface {
	Close()
}

// Iterable is an interface that indicates a type
// can be iterated by a provided Iterator
type Iterable[T any] interface {
	GetIterator() Iterator[T]
}

func AsDefault[T any]() T {
	var t T
	return t
}

// Iterator is an interface that represents a type
// that can be iterated over
type Iterator[T any] interface {
	HasNext() bool
	Value() T
}

type Runnable func()
type Callable[T any] func() T
type ApplyT[T, R any] func(T) R
type AcceptT[T any] func(T)
type BiAccept[T, U any] func(T, U)

type Executor interface {
	Execute(Runnable)
}

type Clearable[T any] interface {
	Clear() Iterator[T]
}

func PrintIfError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func CAS(addr *any, old *any, new *any) bool {
	ptr := unsafe.Pointer(addr)
	return atomic.CompareAndSwapPointer(&ptr, unsafe.Pointer(old), unsafe.Pointer(new))
}
