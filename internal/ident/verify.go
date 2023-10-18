package ident

import (
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type msgSerializer interface {
	SerializeMsg() ([]byte, error)
}

func verify(obj msgSerializer, pubkey []byte, sig *auth.Signature) error {
	msg, err := obj.SerializeMsg()
	if err != nil {
		return err
	}
	return verifySig(pubkey, msg, sig)
}

// VerifyTransaction verifies a transaction's signature using the Authenticator
// registry in this package.
func VerifyTransaction(tx *transactions.Transaction) error {
	return verify(tx, tx.Sender, tx.Signature)
}

// VerifyMessage verifies a message's signature using the Authenticator
// registry in this package.
func VerifyMessage(callMsg *transactions.CallMessage) error {
	return verify(callMsg, callMsg.Sender, callMsg.Signature)
}
