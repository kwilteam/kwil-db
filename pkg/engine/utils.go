package engine

import (
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

// GenerateDBID generates a DBID from a name and owner
func GenerateDBID(name string, owner []byte) string {
	return utils.GenerateDBID(name, owner)
}
