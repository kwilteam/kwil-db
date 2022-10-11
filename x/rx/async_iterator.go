package rx

import "context"

type BoolTask Task[bool]

type AsyncIterator[T any] interface {
	Current() T
	Next(context.Context) BoolTask
	Close() Action
}
