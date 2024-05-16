package http

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	httpTx "github.com/kwilteam/kwil-db/core/rpc/http/tx"
	"github.com/kwilteam/kwil-db/core/types"
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
	if tx == nil {
		return nil, nil
	}

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

func convertToSchema(schema *httpTx.TxSchema) (*types.Schema, error) {
	decodedOwner, err := base64.StdEncoding.DecodeString(schema.Owner)
	if err != nil {
		return nil, err
	}

	s := &types.Schema{
		Owner:             decodedOwner,
		Name:              schema.Name,
		Tables:            convertHttpTables(schema.Tables),
		Actions:           convertHttpActions(schema.Actions),
		Extensions:        convertHttpExtensions(schema.Extensions),
		Procedures:        convertHttpProcedures(schema.Procedures),
		ForeignProcedures: convertHTTPForeignProcedures(schema.ForeignProcedures),
	}

	err = s.Clean()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func convertHttpTables(tables []httpTx.TxTable) []*types.Table {
	tbls := make([]*types.Table, len(tables))

	for i, table := range tables {
		tbls[i] = &types.Table{
			Name:        table.Name,
			Columns:     convertHttpColumns(table.Columns),
			Indexes:     convertHttpIndexes(table.Indexes),
			ForeignKeys: convertHttpForeignKeys(table.ForeignKeys),
		}
	}

	return tbls
}

func convertHttpColumns(columns []httpTx.TxColumn) []*types.Column {
	cols := make([]*types.Column, len(columns))

	for i, column := range columns {
		cols[i] = &types.Column{
			Name:       column.Name,
			Type:       convertDataType(column.Type_),
			Attributes: convertHttpAttributes(column.Attributes),
		}
	}

	return cols
}

func convertDataType(d *httpTx.TxDataType) *types.DataType {
	return &types.DataType{
		Name:    d.Name,
		IsArray: d.IsArray,
	}
}

func convertHttpAttributes(attributes []httpTx.TxAttribute) []*types.Attribute {
	attrs := make([]*types.Attribute, len(attributes))
	for i, attribute := range attributes {
		attrs[i] = &types.Attribute{
			Type:  types.AttributeType(attribute.Type_),
			Value: attribute.Value,
		}
	}
	return attrs
}

func convertHttpIndexes(indexes []httpTx.TxIndex) []*types.Index {
	idxs := make([]*types.Index, len(indexes))
	for i, index := range indexes {
		idxs[i] = &types.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    types.IndexType(index.Type_),
		}
	}
	return idxs
}

func convertHttpForeignKeys(foreignKeys []httpTx.TxForeignKey) []*types.ForeignKey {
	fks := make([]*types.ForeignKey, len(foreignKeys))
	for i, fk := range foreignKeys {
		actions := make([]*types.ForeignKeyAction, len(fk.Actions))
		for j, action := range fk.Actions {
			actions[j] = &types.ForeignKeyAction{
				On: types.ForeignKeyActionOn(action.On),
				Do: types.ForeignKeyActionDo(action.Do),
			}
		}

		fks[i] = &types.ForeignKey{
			ChildKeys:   fk.ChildKeys,
			ParentKeys:  fk.ParentKeys,
			ParentTable: fk.ParentTable,
			Actions:     actions,
		}
	}
	return fks
}

func convertHttpActions(actions []httpTx.TxAction) []*types.Action {
	acts := make([]*types.Action, len(actions))
	for i, action := range actions {
		mods := make([]types.Modifier, len(action.Modifiers))
		for j, mod := range action.Modifiers {
			mods[j] = types.Modifier(mod)
		}

		acts[i] = &types.Action{
			Name:        action.Name,
			Annotations: action.Annotations,
			Parameters:  action.Parameters,
			Public:      action.Public,
			Modifiers:   mods,
			Body:        action.Body,
		}
	}
	return acts
}

func convertHttpExtensions(extensions []httpTx.TxExtensions) []*types.Extension {
	exts := make([]*types.Extension, len(extensions))
	for i, extension := range extensions {
		initialize := make([]*types.ExtensionConfig, len(extension.Initialization))
		for j, init := range extension.Initialization {
			initialize[j] = &types.ExtensionConfig{
				Key:   init.Argument,
				Value: init.Value,
			}
		}

		exts[i] = &types.Extension{
			Name:           extension.Name,
			Initialization: initialize,
			Alias:          extension.Alias,
		}
	}
	return exts
}

func convertHttpProcedures(procedures []httpTx.TxProcedure) []*types.Procedure {
	procs := make([]*types.Procedure, len(procedures))
	for i, procedure := range procedures {
		mods := make([]types.Modifier, len(procedure.Modifiers))
		for j, mod := range procedure.Modifiers {
			mods[j] = types.Modifier(mod)
		}

		var returns *types.ProcedureReturn
		if procedure.ReturnTypes != nil {
			returns = &types.ProcedureReturn{
				IsTable: procedure.ReturnTypes.IsTable,
				Fields:  make([]*types.NamedType, len(procedure.ReturnTypes.Fields)),
			}

			for j, col := range procedure.ReturnTypes.Fields {
				returns.Fields[j] = &types.NamedType{
					Name: col.Name,
					Type: convertDataType(col.Type_),
				}
			}
		}

		procs[i] = &types.Procedure{
			Name:        procedure.Name,
			Annotations: procedure.Annotations,
			Modifiers:   mods,
			Public:      procedure.Public,
			Parameters:  convertHttpParameters(procedure.Parameters),
			Body:        procedure.Body,
		}
	}
	return procs
}

func convertHttpParameters(parameters []httpTx.TxTypedVariable) []*types.ProcedureParameter {
	params := make([]*types.ProcedureParameter, len(parameters))
	for i, param := range parameters {
		params[i] = &types.ProcedureParameter{
			Name: param.Name,
			Type: convertDataType(param.Type_),
		}
	}
	return params
}

func convertHTTPForeignProcedures(procedures []httpTx.TxForeignProcedure) []*types.ForeignProcedure {
	procs := make([]*types.ForeignProcedure, len(procedures))
	for i, procedure := range procedures {
		params := make([]*types.DataType, len(procedure.Parameters))
		for j, param := range procedure.Parameters {
			params[j] = convertDataType(&param)
		}

		var returns *types.ProcedureReturn
		if procedure.ReturnTypes != nil {
			returns = &types.ProcedureReturn{
				IsTable: procedure.ReturnTypes.IsTable,
				Fields:  make([]*types.NamedType, len(procedure.ReturnTypes.Fields)),
			}

			for j, col := range procedure.ReturnTypes.Fields {
				returns.Fields[j] = &types.NamedType{
					Name: col.Name,
					Type: convertDataType(col.Type_),
				}
			}
		}

		procs[i] = &types.ForeignProcedure{
			Name:       procedure.Name,
			Parameters: params,
			Returns:    returns,
		}
	}
	return procs
}
