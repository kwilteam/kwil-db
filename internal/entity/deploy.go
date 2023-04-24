package entity

import (
	"kwil/pkg/engine/models"
	"kwil/pkg/tx"
)

type DeployDatabase struct {
	Schema *models.Dataset
	Tx     *tx.Transaction
}

type DropDatabase struct {
	DBID string
	Tx   *tx.Transaction
}
