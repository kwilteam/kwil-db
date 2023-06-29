package engine

import "context"

type ExtensionInitializer interface {
	CreateInstance(ctx context.Context, metadata map[string]string) (ExtensionInstance, error)
}

type ExtensionInstance interface {
	Execute(ctx context.Context, method string, args ...any) ([]any, error)
}
