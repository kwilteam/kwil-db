package extensions

import (
	"context"
	"fmt"
	"strings"
)

type Extension struct {
	name    string
	url     string
	methods map[string]struct{}

	client ExtensionClient
}

func (e *Extension) Name() string {
	return e.name
}

// New connects to the given extension, and attempts to configure it with the given config.
// If the extension is not available, an error is returned.
func New(url string) *Extension {
	return &Extension{
		name:    "",
		url:     url,
		methods: make(map[string]struct{}),
	}
}

func (e *Extension) Connect(ctx context.Context) error {
	extClient, err := ConnectFunc.Connect(ctx, e.url)
	if err != nil {
		return fmt.Errorf("failed to connect to extension at %s: %w", e.url, err)
	}

	name, err := extClient.GetName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get extension name: %w", err)
	}

	e.name = name
	e.client = extClient

	err = e.loadMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to load methods for extension %s: %w", e.name, err)
	}

	return nil
}

func (e *Extension) loadMethods(ctx context.Context) error {
	methodList, err := e.client.ListMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to list methods for extension '%s' at target '%s': %w", e.name, e.url, err)
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
