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
	newMetadata, err := e.client.Initialize(ctx, metadata)
	if err != nil {
		return nil, err
	}

	return &Instance{
		metadata:   newMetadata,
		extenstion: e,
	}, nil
}

func (i *Instance) Metadata() map[string]string {
	return i.metadata
}

func (i *Instance) Name() string {
	return i.extenstion.name
}

func (i *Instance) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)
	_, ok := i.extenstion.methods[lowerMethod]
	if !ok {
		return nil, fmt.Errorf("method '%s' is not available for extension '%s' at target '%s'", lowerMethod, i.extenstion.name, i.extenstion.url)
	}

	return i.extenstion.client.CallMethod(&types.ExecutionContext{
		Ctx:      ctx,
		Metadata: i.metadata,
	}, lowerMethod, args...)
}
