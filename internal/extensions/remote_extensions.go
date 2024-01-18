package extensions

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-extensions/types"
)

// Remote Extension used for docker extensions defined and deployed remotely
type RemoteExtension struct {
	// Name of the extension
	name string
	// url of the extension server
	url string
	// methods supported by the extension
	methods map[string]struct{}
	// client to connect to the server
	client ExtensionClient
}

func (e *RemoteExtension) Name() string {
	return e.name
}

// New returns a placeholder for the RemoteExtension at a given url
func New(url string) *RemoteExtension {
	return &RemoteExtension{
		name:    "",
		url:     url,
		methods: make(map[string]struct{}),
	}
}

// Initialize initializes based on the given metadata and returns the updated metadata
func (e *RemoteExtension) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	return e.client.Initialize(ctx, metadata)
}

// Execute executes the requested method of an extension. If the method is not supported, an error is returned.
func (e *RemoteExtension) Execute(ctx *execution.ProcedureContext, metadata map[string]string, method string, args ...any) ([]any, error) {
	_, ok := e.methods[method]
	if !ok {
		return nil, fmt.Errorf("method '%s' is not available for extension '%s' at target '%s'", method, e.name, e.url)
	}

	return e.client.CallMethod(&types.ExecutionContext{
		Ctx:      ctx.Ctx,
		Metadata: metadata,
	}, method, args...)
}

type Contexter interface {
	Ctx() context.Context
}

// Connect connects to the given extension, and attempts to configure it with the given config.
// If the extension is not available, an error is returned.
func (e *RemoteExtension) Connect(ctx context.Context) error {
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

func (e *RemoteExtension) loadMethods(ctx context.Context) error {
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
