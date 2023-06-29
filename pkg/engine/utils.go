package engine

import "github.com/kwilteam/kwil-db/pkg/engine/utils"

func GenerateDBID(name, owner string) string {
	return utils.GenerateDBID(name, owner)
}
