package client

import (
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func convertTx(incoming *transactions.Transaction) *txpb.Transaction {
	return &txpb.Transaction{
		Body: &txpb.Transaction_Body{
			Payload:     incoming.Body.Payload,
			PayloadType: incoming.Body.PayloadType.String(),
			Fee:         incoming.Body.Fee.String(),
			Nonce:       incoming.Body.Nonce,
			Salt:        incoming.Body.Salt,
		},
		Signature: convertActionSignature(incoming.Signature),
		Sender:    incoming.Sender,
	}
}

func convertActionSignature(oldSig *crypto.Signature) *txpb.Signature {
	if oldSig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: oldSig.Signature,
		SignatureType:  oldSig.Type.String(),
	}

	return newSig
}
