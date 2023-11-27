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

// VerifySignature verifies the signature given a public key and the message.
// It uses the Authenticator registry in this package.
func VerifySignature(pubkey []byte, msg []byte, sig *auth.Signature) error {
	return verifySig(pubkey, msg, sig)
}
