package common

type Closeable interface {
	Close()
}

type Iterator[T any] interface {
	HasNext() bool
	Value() T
}
