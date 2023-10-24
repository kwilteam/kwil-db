package conversion

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// ConvertFromPBTx converts a protobuf transaction to an abci transaction
func ConvertFromPBTx(incoming *txpb.Transaction) (*transactions.Transaction, error) {
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

// ConvertToPBTx converts an abci transaction(encoded) to a protobuf transaction
func ConvertToPBTx(tx *transactions.Transaction) *txpb.Transaction {
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
		Signature:     ConvertToPBCryptoSignature(tx.Signature),
		Sender:        tx.Sender,
	}
}

// ConvertFromPBCryptoSignature convert a protobuf signature to crypto signature
func ConvertFromPBCryptoSignature(sig *txpb.Signature) *auth.Signature {
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

// ConvertToPBCryptoSignature Convert a crypto signature to protobuf signature
func ConvertToPBCryptoSignature(sig *auth.Signature) *txpb.Signature {
	if sig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: sig.Signature,
		SignatureType:  sig.Type,
	}

	return newSig
}

// TranslateFromPBTxResult convert a protobuf tx result to vanilla tx result
// NOTE: here i try to indicate this `conversion` won't throw error, not sure
// if this is a good idea
func TranslateFromPBTxResult(resp *txpb.TransactionResult) *transactions.TransactionResult {
	return &transactions.TransactionResult{
		Code:      resp.Code,
		Log:       resp.Log,
		GasUsed:   resp.GasUsed,
		GasWanted: resp.GasWanted,
		Data:      resp.Data,
		Events:    resp.Events,
	}
}

func ConvertFromPBTxQueryResp(resp *txpb.TxQueryResponse) (*transactions.TcTxQueryResponse, error) {
	tx, err := ConvertFromPBTx(resp.Tx)
	if err != nil {
		return nil, err
	}

	txResult := TranslateFromPBTxResult(resp.TxResult)

	return &transactions.TcTxQueryResponse{
		Hash:     resp.Hash,
		Height:   resp.Height,
		Tx:       *tx,
		TxResult: *txResult,
	}, nil
}

func ConvertFromPBSchema(dataset *txpb.Schema) *transactions.Schema {
	return &transactions.Schema{
		Owner:   dataset.Owner,
		Name:    dataset.Name,
		Tables:  convertFromPBTables(dataset.Tables),
		Actions: convertFromPBActions(dataset.Actions),
	}
}

func convertFromPBTables(tables []*txpb.Table) []*transactions.Table {
	convTables := make([]*transactions.Table, len(tables))
	for i, table := range tables {
		convTables[i] = &transactions.Table{
			Name:    table.Name,
			Columns: convertFromPBColumns(table.Columns),
			Indexes: convertFromPBIndexes(table.Indexes),
		}
	}

	return convTables
}

func convertFromPBColumns(columns []*txpb.Column) []*transactions.Column {
	convColumns := make([]*transactions.Column, len(columns))
	for i, column := range columns {
		convColumns[i] = &transactions.Column{
			Name:       column.Name,
			Type:       column.Type,
			Attributes: convertFromPBAttributes(column.Attributes),
		}
	}

	return convColumns
}

func convertFromPBAttributes(attributes []*txpb.Attribute) []*transactions.Attribute {
	convAttributes := make([]*transactions.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttributes[i] = &transactions.Attribute{
			Type:  attribute.Type,
			Value: attribute.Value,
		}
	}

	return convAttributes
}

func convertFromPBIndexes(indexes []*txpb.Index) []*transactions.Index {
	convIndexes := make([]*transactions.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = &transactions.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type,
		}
	}

	return convIndexes
}

func convertFromPBActions(actions []*txpb.Action) []*transactions.Action {
	convActions := make([]*transactions.Action, len(actions))
	for i, action := range actions {
		convActions[i] = &transactions.Action{
			Name:       action.Name,
			Public:     action.Public,
			Mutability: action.Mutability,
			Inputs:     action.Inputs,
			Statements: action.Statements,
		}
	}

	return convActions
}
