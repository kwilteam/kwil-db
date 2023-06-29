package extensions

import (
	"context"
	"fmt"
	"strings"
)

type ExtensionConfig struct {
	Name            string            `json:"name"`
	Url             string            `json:"url"`
	ConfigVariables map[string]string `json:"config"`
}

type Extension struct {
	config  *ExtensionConfig
	methods map[string]struct{}

	client ExtensionClient
}

// New connects to the given extension, and attempts to configure it with the given config.
// If the extension is not available, an error is returned.
func New(conf *ExtensionConfig) *Extension { // I return the interface here b/c i think it makes the package api cleaner
	return &Extension{
		config: &ExtensionConfig{
			Name:            conf.Name,
			Url:             conf.Url,
			ConfigVariables: conf.ConfigVariables,
		},
		methods: make(map[string]struct{}),
	}
}

func (e *Extension) Connect(ctx context.Context) error {
	extClient, err := ConnectFunc.Connect(ctx, e.config.Url)
	if err != nil {
		return fmt.Errorf("failed to connect to extension %s: %w", e.config.Name, err)
	}

	err = extClient.Configure(ctx, e.config.ConfigVariables)
	if err != nil {
		return fmt.Errorf("failed to configure extension %s: %w", e.config.Name, err)
	}

	e.client = extClient

	err = e.loadMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to load methods for extension %s: %w", e.config.Name, err)
	}

	return nil
}

func (e *Extension) loadMethods(ctx context.Context) error {
	methodList, err := e.client.ListMethods(ctx)
	if err != nil {
		return fmt.Errorf("failed to list methods for extension %s: %w", e.config.Name, err)
	}

	e.methods = make(map[string]struct{})
	for _, method := range methodList {
		lowerName := strings.ToLower(method)

		_, ok := e.methods[lowerName]
		if ok {
			return fmt.Errorf("extension %s has duplicate method %s. this is an issue with the extension", e.config.Name, lowerName)
		}

		e.methods[lowerName] = struct{}{}
	}

	return nil
}
