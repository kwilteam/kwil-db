package dataset3

import "github.com/kwilteam/kwil-db/pkg/engine/utils"

func generateDBID(name, owner string) string {
	return utils.GenerateDBID(name, owner)
}
