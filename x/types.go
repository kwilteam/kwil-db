package x

func init() {
	ch := make(chan Void)
	close(ch)
	ClosedChan = ch
}

// Void An empty struct{} typed as 'Void'
type Void struct{}

// ClosedChan is a closed Void channel that can be
// used in default closed use cases or in context
// of atomic CAS operations for closure scenarios.
var ClosedChan chan Void

// Tuple2 is a combination of two values of any type.
// In general, it is used in place of declaring one-off
// structs for passing around a pair of values.
type Tuple2[T, U any] struct {
	First  T
	Second U
}

// Tuple3 is a combination of three values of any type.
// In general, it is used in place of declaring one-off
// structs for passing around three values.
type Tuple3[T, U, V any] struct {
	First  T
	Second U
	Third  V
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

// Iterator is an interface that represents a type
// that can be iterated over
type Iterator[T any] interface {
	HasNext() bool
	Value() T
}

func AsDefault[T any]() T {
	var t T
	return t
}
