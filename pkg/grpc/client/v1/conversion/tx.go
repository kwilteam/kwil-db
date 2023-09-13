package conversion

import (
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/client/types"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/transactions"
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

	convSignature, err := ConvertToCryptoSignature(incoming.Signature)
	if err != nil {
		return nil, err
	}

	bigFee, ok := big.NewInt(0).SetString(incoming.Body.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("invalid fee: %s", incoming.Body.Fee)
	}

	return &transactions.Transaction{
		Body: &transactions.TransactionBody{
			PayloadType: payloadType,
			Payload:     incoming.Body.Payload,
			Nonce:       incoming.Body.Nonce,
			Fee:         bigFee,
			Salt:        incoming.Body.Salt,
			Description: incoming.Body.Description,
		},
		Serialization: transactions.SignedMsgSerializationType(incoming.Serialization),
		Signature:     convSignature,
		Sender:        incoming.Sender,
	}, nil
}

// ConvertFromAbciTx converts an abci transaction(encoded) to a protobuf transaction
func ConvertFromAbciTx(tx *transactions.Transaction) (*txpb.Transaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}

	return &txpb.Transaction{
		Body: &txpb.Transaction_Body{
			Payload:     tx.Body.Payload,
			PayloadType: tx.Body.PayloadType.String(),
			Fee:         tx.Body.Fee.String(),
			Nonce:       tx.Body.Nonce,
			Salt:        tx.Body.Salt,
			Description: tx.Body.Description,
		},
		Serialization: tx.Serialization.String(),
		Signature:     ConvertFromCryptoSignature(tx.Signature),
		Sender:        tx.Sender,
	}, nil
}

func newEmptySignature() (bytes []byte, sigType crypto.SignatureType) {
	return []byte{}, crypto.SignatureTypeEmpty
}

// ConvertToCryptoSignature convert a protobuf signature to crypto signature
func ConvertToCryptoSignature(sig *txpb.Signature) (*crypto.Signature, error) {
	if sig == nil {
		emptyBts, emptyType := newEmptySignature()
		return &crypto.Signature{
			Signature: emptyBts,
			Type:      emptyType,
		}, nil
	}

	sigType := crypto.SignatureTypeLookUp(sig.SignatureType)
	return &crypto.Signature{
		Signature: sig.SignatureBytes,
		Type:      sigType,
	}, nil
}

// ConvertFromCryptoSignature Convert a crypto signature to protobuf signature
func ConvertFromCryptoSignature(sig *crypto.Signature) *txpb.Signature {
	if sig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: sig.Signature,
		SignatureType:  sig.Type.String(),
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

func ConvertToTxQueryResp(resp *txpb.TxQueryResponse) (*types.TcTxQueryResponse, error) {
	tx, err := ConvertToAbciTx(resp.Tx)
	if err != nil {
		return nil, err
	}

	txResult := TranslateToTxResult(resp.TxResult)

	return &types.TcTxQueryResponse{
		Hash:     resp.Hash,
		Height:   resp.Height,
		Tx:       *tx,
		TxResult: *txResult,
	}, nil
}
