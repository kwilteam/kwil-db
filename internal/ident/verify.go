package ident

import (
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// VerifyTransaction verifies a transaction's signature using the Authenticator
// registry in this package.
func VerifyTransaction(chainID string, tx *transactions.Transaction) error {
	msg, err := tx.SerializeMsg(chainID)
	if err != nil {
		return err
	}
	return verifySig(tx.Sender, msg, tx.Signature)
}

// VerifyMessage verifies a message's signature using the Authenticator
// registry in this package.
func VerifyMessage(callMsg *transactions.CallMessage) error {
	msg, err := callMsg.SerializeMsg()
	if err != nil {
		return err
	}
	return verifySig(callMsg.Sender, msg, callMsg.Signature)
}
