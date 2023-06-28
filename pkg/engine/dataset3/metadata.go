package dataset3

import "github.com/kwilteam/kwil-db/pkg/engine/types"

type Metadata struct {
	Owner      string
	Name       string
	Procedures map[string]*types.Procedure
}

func (m *Metadata) DBID() string {
	return generateDBID(m.Name, m.Owner)
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
