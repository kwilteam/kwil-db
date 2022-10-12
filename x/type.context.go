package x

import "fmt"
import "context"

var root_context context.Context

func init() {
	// use for context propagation needs (e.g. tracing, logging,
	// IoC/DI access, etc). Initial usage for creation was for passing
	// the producer in a request context to a gRpc db service handler.
	root_context = context.Background()
}

var errContextIdEmpty = fmt.Errorf("lookup id cannot be empty")

func ErrContextIdEmpty() error {
	return errContextIdEmpty
}

// RootContext returns the root context.
func RootContext() context.Context {
	return root_context
}

// Unwrap will unwrap the value from context given id and value. If
// the context is nil, the golang default for type T will be returned.
// If the id is empty, the method will panic.
func Unwrap[T interface{}](ctx context.Context, id string) (out T) {
	if id == "" {
		return
	}

	e, ok := ctx.Value(id).(T)
	if !ok {
		return
	}

	return e
}

// UnwrapWithDefault will unwrap the value from context given id and value.
// If the context is nil, the defaultValue param will be returned. If the
// id is empty, the method will panic.
func UnwrapWithDefault[T any](ctx context.Context, id string, defaultValue T) T {
	if id == "" {
		panic(errContextIdEmpty)
	}

	e, ok := ctx.Value(id).(T)
	if !ok {
		return defaultValue
	}

	return e
}

// Wrap will wrap the context with the given id and value. If the context
// is nil, then it will use the root context. If the id is empty, the
// method will panic.
func Wrap[T any](ctx context.Context, id string, item T) context.Context {
	if id == "" {
		panic(errContextIdEmpty)
	}

	if ctx == nil {
		ctx = root_context
	}

	return context.WithValue(ctx, id, item)
}
