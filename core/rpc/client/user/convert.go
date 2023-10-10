package client

import (
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func convertTx(incoming *transactions.Transaction) *txpb.Transaction {
	return conversion.ConvertFromAbciTx(incoming)
}

func convertActionSignature(oldSig *auth.Signature) *txpb.Signature {
	return conversion.ConvertFromCryptoSignature(oldSig)
}
