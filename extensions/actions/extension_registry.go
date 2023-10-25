package extensions

import (
	"context"
	"strings"
)

type Extension interface {
	Name() string
	Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error)
	Execute(ctx context.Context, metadata map[string]string, method string, args ...any) ([]any, error)
}

var registeredExtensions = make(map[string]Extension)

func RegisterExtension(name string, ext Extension) error {
	name = strings.ToLower(name)
	if _, ok := registeredExtensions[name]; ok {
		panic("extension of same name already registered: " + name)
	}

	registeredExtensions[name] = ext
	return nil
}

func RegisteredExtensions() map[string]Extension {
	return registeredExtensions
}
