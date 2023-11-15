package http2

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	httpTx "github.com/kwilteam/kwil-db/core/rpc/http/tx"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// convertTx converts a transaction to a httpTx.TxTransaction
func convertTx(tx *transactions.Transaction) *httpTx.TxTransaction {
	if tx.Sender == nil {
		tx.Sender = []byte{}
	}
	if tx.Signature == nil {
		tx.Signature = &auth.Signature{
			Signature: []byte{},
			Type:      "",
		}
	}

	return &httpTx.TxTransaction{
		Sender:        base64.StdEncoding.EncodeToString(tx.Sender),
		Serialization: tx.Serialization.String(),
		Signature: &httpTx.TxSignature{
			SignatureBytes: base64.StdEncoding.EncodeToString(tx.Signature.Signature),
			SignatureType:  tx.Signature.Type,
		},
		Body: &httpTx.TxTransactionBody{
			Payload:     base64.StdEncoding.EncodeToString(tx.Body.Payload),
			PayloadType: tx.Body.PayloadType.String(),
			Fee:         tx.Body.Fee.String(),
			Nonce:       strconv.FormatUint(tx.Body.Nonce, 10),
			ChainId:     tx.Body.ChainID,
			Description: tx.Body.Description,
		},
	}
}

// convertHttpTx converts a httpTx.TxTransaction to a transactions.Transaction
func convertHttpTx(tx *httpTx.TxTransaction) (*transactions.Transaction, error) {
	decodedSender, err := base64.StdEncoding.DecodeString(tx.Sender)
	if err != nil {
		return nil, err
	}

	decodedSignature, err := base64.StdEncoding.DecodeString(tx.Signature.SignatureBytes)
	if err != nil {
		return nil, err
	}

	decodedPayload, err := base64.StdEncoding.DecodeString(tx.Body.Payload)
	if err != nil {
		return nil, err
	}

	fee, ok := new(big.Int).SetString(tx.Body.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse fee to big.Int. received: %s", tx.Body.Fee)
	}

	decodedNonce, err := strconv.ParseUint(tx.Body.Nonce, 10, 64)
	if err != nil {
		return nil, err
	}

	return &transactions.Transaction{
		Sender:        decodedSender,
		Serialization: transactions.SignedMsgSerializationType(tx.Serialization),
		Signature: &auth.Signature{
			Signature: decodedSignature,
			Type:      tx.Signature.SignatureType,
		},
		Body: &transactions.TransactionBody{
			Payload:     decodedPayload,
			PayloadType: transactions.PayloadType(tx.Body.PayloadType),
			Fee:         fee,
			Nonce:       decodedNonce,
			ChainID:     tx.Body.ChainId,
			Description: tx.Body.Description,
		},
	}, nil
}

// convertHttpTxResult converts a httpTx.TxTransactionResult to a transactions.TransactionResult
func convertHttpTxResult(result *httpTx.TxTransactionResult) (*transactions.TransactionResult, error) {
	decodedGasUsed, err := strconv.ParseInt(result.GasUsed, 10, 64)
	if err != nil {
		return nil, err
	}

	decodedGasWanted, err := strconv.ParseInt(result.GasWanted, 10, 64)
	if err != nil {
		return nil, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return nil, err
	}

	decodedEvents := make([][]byte, 0, len(result.Events))
	for _, event := range result.Events {
		decodedEvent, err := base64.StdEncoding.DecodeString(event)
		if err != nil {
			return nil, err
		}
		decodedEvents = append(decodedEvents, decodedEvent)
	}

	return &transactions.TransactionResult{
		Code:      uint32(result.Code),
		Log:       result.Log,
		GasUsed:   decodedGasUsed,
		GasWanted: decodedGasWanted,
		Data:      decodedData,
		Events:    decodedEvents,
	}, nil
}

// convertHttpSchema converts a httpTx.TxSchema to a transactions.Schema
func convertHttpSchema(schema *httpTx.TxSchema) (*transactions.Schema, error) {
	decodedOwner, err := base64.StdEncoding.DecodeString(schema.Owner)
	if err != nil {
		return nil, err
	}

	return &transactions.Schema{
		Owner:      decodedOwner,
		Name:       schema.Name,
		Tables:     convertHttpTables(schema.Tables),
		Actions:    convertHttpActions(schema.Actions),
		Extensions: convertHttpExtensions(schema.Extensions),
	}, nil
}

// convertHttpTables converts []httpTx.TxTable to a []transactions.Table
func convertHttpTables(tables []httpTx.TxTable) []*transactions.Table {
	tbls := make([]*transactions.Table, len(tables))

	for i, table := range tables {
		tbls[i] = &transactions.Table{
			Name:        table.Name,
			Columns:     convertHttpColumns(table.Columns),
			Indexes:     convertHttpIndexes(table.Indexes),
			ForeignKeys: convertHttpForeignKeys(table.ForeignKeys),
		}
	}

	return tbls
}

// convertHttpColumns converts []httpTx.TxColumn to []transactions.Column
func convertHttpColumns(columns []httpTx.TxColumn) []*transactions.Column {
	cols := make([]*transactions.Column, len(columns))

	for i, column := range columns {
		cols[i] = &transactions.Column{
			Name:       column.Name,
			Type:       column.Type_,
			Attributes: convertHttpAttributes(column.Attributes),
		}
	}

	return cols
}

// convertHttpAttributes converts []httpTx.TxAttribute to []transactions.Attribute
func convertHttpAttributes(attributes []httpTx.TxAttribute) []*transactions.Attribute {
	attrs := make([]*transactions.Attribute, len(attributes))
	for i, attribute := range attributes {
		attrs[i] = &transactions.Attribute{
			Type:  attribute.Type_,
			Value: attribute.Value,
		}
	}
	return attrs
}

// convertHttpIndexes converts []httpTx.TxIndex to []transactions.Index
func convertHttpIndexes(indexes []httpTx.TxIndex) []*transactions.Index {
	idxs := make([]*transactions.Index, len(indexes))
	for i, index := range indexes {
		idxs[i] = &transactions.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type_,
		}
	}
	return idxs
}

// convertHttpForeignKeys converts []httpTx.TxForeignKey to []transactions.ForeignKey
func convertHttpForeignKeys(foreignKeys []httpTx.TxForeignKey) []*transactions.ForeignKey {
	fks := make([]*transactions.ForeignKey, len(foreignKeys))
	for i, fk := range foreignKeys {
		actions := make([]*transactions.ForeignKeyAction, len(fk.Actions))
		for j, action := range fk.Actions {
			actions[j] = &transactions.ForeignKeyAction{
				On: action.On,
				Do: action.Do,
			}
		}

		fks[i] = &transactions.ForeignKey{
			ChildKeys:   fk.ChildKeys,
			ParentKeys:  fk.ParentKeys,
			ParentTable: fk.ParentTable,
			Actions:     actions,
		}
	}
	return fks
}

// convertHttpActions converts []httpTx.TxAction to []transactions.Action
func convertHttpActions(actions []httpTx.TxAction) []*transactions.Action {
	acts := make([]*transactions.Action, len(actions))
	for i, action := range actions {
		acts[i] = &transactions.Action{
			Name:        action.Name,
			Annotations: action.Annotations,
			Inputs:      action.Inputs,
			Mutability:  action.Mutability,
			Auxiliaries: action.Auxiliaries,
			Public:      action.Public,
			Statements:  action.Statements,
		}
	}
	return acts
}

// convertHttpExtensions converts []httpTx.TxExtension to []transactions.Extension
func convertHttpExtensions(extensions []httpTx.TxExtensions) []*transactions.Extension {
	exts := make([]*transactions.Extension, len(extensions))
	for i, extension := range extensions {
		initialize := make([]*transactions.ExtensionConfig, len(extension.Initialization))
		for j, init := range extension.Initialization {
			initialize[j] = &transactions.ExtensionConfig{
				Argument: init.Argument,
				Value:    init.Value,
			}
		}

		exts[i] = &transactions.Extension{
			Name:   extension.Name,
			Config: initialize,
			Alias:  extension.Alias,
		}
	}
	return exts
}
