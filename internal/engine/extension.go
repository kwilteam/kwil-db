package engine

import "context"

// An ExtensionInitializer is a function that calls an extension, creating an instance of it.
type ExtensionInitializer interface {
	CreateInstance(ctx context.Context, metadata map[string]string) (ExtensionInstance, error)
}

// An ExtensionInstance is an instance of an extension.
type ExtensionInstance interface {
	Execute(ctx context.Context, method string, args ...any) ([]any, error)
}
