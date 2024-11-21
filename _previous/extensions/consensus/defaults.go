package consensus

import (
	"github.com/kwilteam/kwil-db/common/chain/forks"
)

// Register the canonical (non-extension) hard forks that are baked into kwild.
func init() {
	RegisterHardfork(&Hardfork{
		// "halt" is a canonical fork that has a named field in the Forks struct.
		// There may be specialized code anywhere using forks.IsHalt(height).
		// This one includes no standard updates e.g. payloads.
		Name: forks.ForkHalt,

		// NOTE: canonical forks can define any of the standardized updates, but
		// this one does not.
	})
}
