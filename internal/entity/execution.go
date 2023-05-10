package entity

import (
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type ExecuteAction struct {
	Tx            *tx.Transaction
	ExecutionBody *models.ActionExecution
}
