package dataset

import "github.com/kwilteam/kwil-db/pkg/engine/utils"

func (d *Dataset) DBID() string {
	return utils.GenerateDBID(d.name, d.owner.PubKey())
}
