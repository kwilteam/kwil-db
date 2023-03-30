package entity

import "kwil/pkg/tx"

type ExecuteAction struct {
	DBID   string `json:"db_id"`
	Action string `json:"action"`
	Params []map[string][]byte
	Tx     *tx.Transaction
}
