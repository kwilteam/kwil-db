package extensions

import (
	"context"
	"fmt"
	"strings"
)

type Extension interface {
	Connect(ctx context.Context) error
	CreateInstance(ctx context.Context, metadata map[string]string) (*Instance, error)
}

type extension struct {
	name     string
	endpoint string
	config   map[string]string
	methods  map[string]struct{}

	client ExtensionClient
}

// New connects to the given extension, and attempts to configure it with the given config.
// If the extension is not available, an error is returned.
func New(name, endpoint string, config map[string]string) Extension { // I return the interface here b/c i think it makes the package api cleaner
	return &extension{
		name:     name,
		endpoint: endpoint,
		config:   config,
	}
}

func (e *extension) Connect(ctx context.Context) error {
	extClient, err := ConnectFunc.Connect(ctx, e.endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to extension %s: %w", e.name, err)
	}

	err = extClient.Configure(ctx, e.config)
	if err != nil {
		return fmt.Errorf("failed to configure extension %s: %w", e.name, err)
	}

	e.client = extClient

	err = e.loadMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to load methods for extension %s: %w", e.name, err)
	}

	return nil
}

func (e *extension) loadMethods(ctx context.Context) error {
	methodList, err := e.client.ListMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to list methods for extension %s: %w", e.name, err)
	}

	e.methods = make(map[string]struct{})
	for _, method := range methodList {
		lowerName := strings.ToLower(method)

		_, ok := e.methods[lowerName]
		if ok {
			return fmt.Errorf("extension %s has duplicate method %s. this is an issue with the extension", e.name, lowerName)
		}

		e.methods[lowerName] = struct{}{}
	}

	return nil
}
