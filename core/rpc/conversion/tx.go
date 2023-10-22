package conversion

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// ConvertToAbciTx converts a protobuf transaction to an abci transaction
func ConvertToAbciTx(incoming *txpb.Transaction) (*transactions.Transaction, error) {
	payloadType := transactions.PayloadType(incoming.Body.PayloadType)
	if !payloadType.Valid() {
		return nil, fmt.Errorf("invalid payload type: %s", incoming.Body.PayloadType)
	}

	if incoming.Signature == nil {
		return nil, fmt.Errorf("transaction signature not given")
	}

	convSignature := &auth.Signature{
		Signature: incoming.Signature.SignatureBytes,
		Type:      incoming.Signature.SignatureType,
	}

	bigFee, ok := big.NewInt(0).SetString(incoming.Body.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("invalid fee: %s", incoming.Body.Fee)
	}

	return &transactions.Transaction{
		Body: &transactions.TransactionBody{
			Description: incoming.Body.Description,
			PayloadType: payloadType,
			Payload:     incoming.Body.Payload,
			Nonce:       incoming.Body.Nonce,
			Fee:         bigFee,
			ChainID:     incoming.Body.ChainId,
		},
		Serialization: transactions.SignedMsgSerializationType(incoming.Serialization),
		Signature:     convSignature,
		Sender:        incoming.Sender,
	}, nil
}

// ConvertFromAbciTx converts an abci transaction(encoded) to a protobuf transaction
func ConvertFromAbciTx(tx *transactions.Transaction) *txpb.Transaction {
	return &txpb.Transaction{
		Body: &txpb.Transaction_Body{
			Payload:     tx.Body.Payload,
			PayloadType: tx.Body.PayloadType.String(),
			Fee:         tx.Body.Fee.String(),
			Nonce:       tx.Body.Nonce,
			ChainId:     tx.Body.ChainID,
			Description: tx.Body.Description,
		},
		Serialization: tx.Serialization.String(),
		Signature:     ConvertFromCryptoSignature(tx.Signature),
		Sender:        tx.Sender,
	}
}

// ConvertToCryptoSignature convert a protobuf signature to crypto signature
func ConvertToCryptoSignature(sig *txpb.Signature) *auth.Signature {
	// signatures can be empty and still be valid for calls
	if sig == nil {
		return &auth.Signature{
			Signature: []byte{},
			Type:      "",
		}
	}

	return &auth.Signature{
		Signature: sig.SignatureBytes,
		Type:      sig.SignatureType,
	}
}

// ConvertFromCryptoSignature Convert a crypto signature to protobuf signature
func ConvertFromCryptoSignature(sig *auth.Signature) *txpb.Signature {
	if sig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: sig.Signature,
		SignatureType:  sig.Type,
	}

	return newSig
}

// TranslateToTxResult convert a protobuf tx result to vanilla tx result
// NOTE: here i try to indicate this `conversion` won't throw error, not sure
// if this is a good idea
func TranslateToTxResult(resp *txpb.TransactionResult) *transactions.TransactionResult {
	return &transactions.TransactionResult{
		Code:      resp.Code,
		Log:       resp.Log,
		GasUsed:   resp.GasUsed,
		GasWanted: resp.GasWanted,
		Data:      resp.Data,
		Events:    resp.Events,
	}
}

func ConvertToTxQueryResp(resp *txpb.TxQueryResponse) (*transactions.TcTxQueryResponse, error) {
	tx, err := ConvertToAbciTx(resp.Tx)
	if err != nil {
		return nil, err
	}

	txResult := TranslateToTxResult(resp.TxResult)

	return &transactions.TcTxQueryResponse{
		Hash:     resp.Hash,
		Height:   resp.Height,
		Tx:       *tx,
		TxResult: *txResult,
	}, nil
}
