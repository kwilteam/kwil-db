package entity

import (
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type ExecuteAction struct {
	Tx            *tx.Transaction
	ExecutionBody *tx.ExecuteActionPayload
}

// CallAction is a struct that represents the action call
// a call is a read-only action
type CallAction struct {
	Message *tx.SignedMessage[tx.JsonPayload]
	Payload *tx.CallActionPayload
}
