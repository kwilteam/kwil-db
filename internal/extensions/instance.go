package extensions

import (
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
)

// An instance is a single instance of an extension.
// Each Kuneiform schema that uses an extension will have its own instance.
// The instance is a way to encapsulate metadata.
// For example, the instance may contain the smart contract address for an ERC20 token
// that is used by the Kuneiform schema.
type Instance struct {
	metadata map[string]string

	extension LegacyEngineExtension
}

func (i *Instance) Metadata() map[string]string {
	return i.metadata
}

func (i *Instance) Execute(ctx *execution.ProcedureContext, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)
	return i.extension.Execute(ctx, i.metadata, lowerMethod, args...)
}
