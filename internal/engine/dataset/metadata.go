package dataset

import "github.com/kwilteam/kwil-db/internal/engine/types"

type Metadata struct {
	Procedures map[string]*types.Procedure
}

func newMetadata(procs []*types.Procedure) *Metadata {
	metadata := &Metadata{
		Procedures: make(map[string]*types.Procedure),
	}

	for _, proc := range procs {
		metadata.Procedures[proc.Name] = proc
	}

	return metadata
}
