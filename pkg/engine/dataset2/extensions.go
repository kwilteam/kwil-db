package dataset2

import "context"

type InitializedExtension interface {
	Execute(ctx context.Context, method string, args ...any) ([]any, error)
	Metadata() map[string]string
}

type Initializer interface {
	Initialize(context.Context, map[string]string) (InitializedExtension, error)
}
