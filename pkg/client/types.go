package client

import "github.com/kwilteam/kwil-db/pkg/engine/utils"

// some of these mimic internal/entity

type datasetIdentifier struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (d *datasetIdentifier) Dbid() string {
	return utils.GenerateDBID(d.Name, d.Owner)
}

type actionExecution struct {
	Action string           `json:"action"`
	DBID   string           `json:"dbid"`
	Params []map[string]any `json:"params"`
}
