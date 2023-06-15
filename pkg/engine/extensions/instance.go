package extensions

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-extensions/types"
)

// An instance is a single instance of an extension.
// Each Kuneiform schema that uses an extension will have its own instance.
// The instance is a way to encapsulate metadata.
// For example, the instance may contain the smart contract address for an ERC20 token
// that is used by the Kuneiform schema.
type Instance struct {
	metadata map[string]string

	extenstion *Extension
}

func (e *Extension) CreateInstance(ctx context.Context, metadata map[string]string) (*Instance, error) {
	requiredMetadata, err := e.client.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	mergedMetadata, err := mergeMetadata(requiredMetadata, metadata)
	if err != nil {
		return nil, err
	}

	return &Instance{
		metadata:   mergedMetadata,
		extenstion: e,
	}, nil
}

func (i *Instance) Metadata() map[string]string {
	return i.metadata
}

func (i *Instance) Name() string {
	return i.extenstion.name
}

func mergeMetadata(required, provided map[string]string) (map[string]string, error) {
	merged := make(map[string]string)

	for key, defaultValue := range required {
		userValue, ok := provided[key]
		if !ok {
			if defaultValue == "" {
				return nil, fmt.Errorf("missing required metadata %s", key)
			}
			merged[key] = defaultValue
		} else {
			merged[key] = userValue
		}
	}

	return merged, nil
}

func (i *Instance) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)
	_, ok := i.extenstion.methods[lowerMethod]
	if !ok {
		return nil, fmt.Errorf("method %s is not available for extension %s", lowerMethod, i.extenstion.name)
	}

	return i.extenstion.client.CallMethod(&types.ExecutionContext{
		Ctx:      ctx,
		Metadata: i.metadata,
	}, lowerMethod, args...)
}
