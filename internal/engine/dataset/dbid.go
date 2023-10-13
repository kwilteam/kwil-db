package dataset

import "github.com/kwilteam/kwil-db/core/utils"

func (d *Dataset) DBID() string {
	return utils.GenerateDBID(d.name, d.owner.PubKey())
}
