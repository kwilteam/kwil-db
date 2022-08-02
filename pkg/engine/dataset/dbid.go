package dataset

import "github.com/kwilteam/kwil-db/pkg/engine/utils"

func generateDBID(name, owner string) string {
	return utils.GenerateDBID(name, owner)
}

func (d *Dataset) DBID() string {
	return generateDBID(d.name, d.owner)
}
