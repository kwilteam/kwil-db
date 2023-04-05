package entity

import (
	"kwil/pkg/engine/models"
	"kwil/pkg/tx"
)

type ExecuteAction struct {
	Tx            *tx.Transaction
	ExecutionBody *models.ActionExecution
}
