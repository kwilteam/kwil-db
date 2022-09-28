package types

// Integer is a type that represents the various
// uint and int golang types
type Integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Closeable is an interface that represents a type that can be closed
type Closeable interface {
	Close()
}

// Iterator is an interface that represents a type that can be iterated over
type Iterator[T any] interface {
	HasNext() bool
	Value() T
}

type Service interface{}

type ClosableService interface {
	Service
	Closeable
}

type BackgroundService interface {
	Service

	// Start the service.
	Start() error

	// Shutdown the service.
	Shutdown()

	// IsRunning returns the service's status.
	IsRunning() bool

	// AwaitShutdown waits for service shutdown.
	AwaitShutdown()
}
