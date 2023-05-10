package entity

import (
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type DeployDatabase struct {
	Schema *models.Dataset
	Tx     *tx.Transaction
}

type DropDatabase struct {
	DBID string
	Tx   *tx.Transaction
}
