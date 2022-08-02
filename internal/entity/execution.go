package entity

import (
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type ExecuteAction struct {
	Tx            *tx.Transaction
	ExecutionBody *ActionExecution
}

type ActionExecution struct {
	Action string           `json:"action"`
	DBID   string           `json:"dbid"`
	Params []map[string]any `json:"params"`
}
