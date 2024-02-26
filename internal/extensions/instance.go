package extensions

import (
	"context"
	"strings"

	extensions "github.com/kwilteam/kwil-db/extensions/actions"
)

// An instance is a single instance of an extension.
// Each Kuneiform schema that uses an extension will have its own instance.
// The instance is a way to encapsulate metadata.
// For example, the instance may contain the smart contract address for an ERC20 token
// that is used by the Kuneiform schema.
type Instance struct {
	metadata map[string]string

	extension extensions.Extension
}

func (i *Instance) Metadata() map[string]string {
	return i.metadata
}

func (i *Instance) Name() string {
	return i.extension.Name()
}

func (i *Instance) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)
	return i.extension.Execute(ctx, i.metadata, lowerMethod, args...)
}
