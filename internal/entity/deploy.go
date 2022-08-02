package entity

import (
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type DeployDatabase struct {
	Schema *Schema
	Tx     *tx.Transaction
}

type DropDatabase struct {
	DBID string
	Tx   *tx.Transaction
}
