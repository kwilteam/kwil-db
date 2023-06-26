package dataset3

import "github.com/kwilteam/kwil-db/pkg/engine/types"

type Metadata struct {
	Procedures map[string]*types.Procedure
}
