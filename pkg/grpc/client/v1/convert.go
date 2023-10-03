package client

import (
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

// TODO: this is actually duplicated with internal/controller/grpc/txsvc/v1/convert.go
// maybe we should move tx conversion utils functions to pkg/transactions?
func convertTx(incoming *transactions.Transaction) *txpb.Transaction {
	return &txpb.Transaction{
		Body: &txpb.Transaction_Body{
			Description: incoming.Body.Description,
			Payload:     incoming.Body.Payload,
			PayloadType: incoming.Body.PayloadType.String(),
			Fee:         incoming.Body.Fee.String(),
			Nonce:       incoming.Body.Nonce,
			Salt:        incoming.Body.Salt,
		},
		Serialization: incoming.Serialization.String(),
		Signature:     convertActionSignature(incoming.Signature),
		Sender:        incoming.Sender,
	}
}

func convertActionSignature(oldSig *auth.Signature) *txpb.Signature {
	if oldSig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: oldSig.Signature,
		SignatureType:  oldSig.Type,
	}

	return newSig
}
