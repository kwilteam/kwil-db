package grpc

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func convertTx(incoming *transactions.Transaction) *txpb.Transaction {
	return ConvertToPBTx(incoming)
}

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
		Signature:     convertToPBCryptoSignature(tx.Signature),
		Sender:        tx.Sender,
	}
}

// convertToPBCryptoSignature Convert a crypto signature to protobuf signature
func convertToPBCryptoSignature(sig *auth.Signature) *txpb.Signature {
	if sig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: sig.Signature,
		SignatureType:  sig.Type,
	}

	return newSig
}

// translateFromPBTxResult convert a protobuf tx result to vanilla tx result
// NOTE: here i try to indicate this `conversion` won't throw error, not sure
// if this is a good idea
func translateFromPBTxResult(resp *txpb.TransactionResult) *transactions.TransactionResult {
	return &transactions.TransactionResult{
		Code:      resp.Code,
		Log:       resp.Log,
		GasUsed:   resp.GasUsed,
		GasWanted: resp.GasWanted,
		Data:      resp.Data,
		Events:    resp.Events,
	}
}

func convertFromPBTxQueryResp(resp *txpb.TxQueryResponse) (*transactions.TcTxQueryResponse, error) {
	tx, err := ConvertFromPBTx(resp.Tx)
	if err != nil {
		return nil, err
	}

	txResult := translateFromPBTxResult(resp.TxResult)

	return &transactions.TcTxQueryResponse{
		Hash:     resp.Hash,
		Height:   resp.Height,
		Tx:       *tx,
		TxResult: *txResult,
	}, nil
}

func convertPBToSchema(schema *txpb.Schema) (*types.Schema, error) {
	s := &types.Schema{
		Owner:      schema.Owner,
		Name:       schema.Name,
		Tables:     convertTablesToEngine(schema.Tables),
		Actions:    convertActionsToEngine(schema.Actions),
		Extensions: convertExtensionsToEngine(schema.Extensions),
		Procedures: convertProceduresToEngine(schema.Procedures),
	}

	return s, s.Clean()
}

func convertTablesToEngine(tables []*txpb.Table) []*types.Table {
	convTables := make([]*types.Table, len(tables))
	for i, table := range tables {
		convTable := &types.Table{
			Name:        table.Name,
			Columns:     convertColumnsToEngine(table.Columns),
			Indexes:     convertIndexesToEngine(table.Indexes),
			ForeignKeys: convertForeignKeysToEngine(table.ForeignKeys),
		}
		convTables[i] = convTable
	}

	return convTables
}

func convertColumnsToEngine(columns []*txpb.Column) []*types.Column {
	convColumns := make([]*types.Column, len(columns))
	for i, column := range columns {
		convColumn := &types.Column{
			Name:       column.Name,
			Type:       convertDataTypeToEngine(column.Type),
			Attributes: convertAttributesToEngine(column.Attributes),
		}
		convColumns[i] = convColumn
	}

	return convColumns
}

func convertDataTypeToEngine(dataType *txpb.DataType) *types.DataType {
	return &types.DataType{
		Name:    dataType.Name,
		IsArray: dataType.IsArray,
	}
}

func convertAttributesToEngine(attributes []*txpb.Attribute) []*types.Attribute {
	convAttributes := make([]*types.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttribute := &types.Attribute{
			Type:  types.AttributeType(attribute.Type),
			Value: attribute.Value,
		}
		convAttributes[i] = convAttribute
	}

	return convAttributes
}

func convertIndexesToEngine(indexes []*txpb.Index) []*types.Index {
	convIndexes := make([]*types.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = &types.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    types.IndexType(index.Type),
		}
	}

	return convIndexes
}

func convertActionsToEngine(actions []*txpb.Action) []*types.Action {
	convActions := make([]*types.Action, len(actions))
	for i, action := range actions {
		convActions[i] = &types.Action{
			Name:        action.Name,
			Public:      action.Public,
			Parameters:  action.Parameters,
			Modifiers:   convertModifiersToEngine(action.Modifiers),
			Annotations: action.Annotations,
			Body:        action.Body,
		}
	}

	return convActions
}

func convertModifiersToEngine(mods []string) []types.Modifier {
	convModifiers := make([]types.Modifier, len(mods))
	for i, mod := range mods {
		convModifiers[i] = types.Modifier(mod)
	}

	return convModifiers
}

func convertForeignKeysToEngine(foreignKeys []*txpb.ForeignKey) []*types.ForeignKey {
	convForeignKeys := make([]*types.ForeignKey, len(foreignKeys))
	for i, foreignKey := range foreignKeys {
		convertedActions := make([]*types.ForeignKeyAction, len(foreignKey.Actions))

		for j, action := range foreignKey.Actions {
			convertedActions[j] = &types.ForeignKeyAction{
				On: types.ForeignKeyActionOn(action.On),
				Do: types.ForeignKeyActionDo(action.Do),
			}
		}

		convForeignKeys[i] = &types.ForeignKey{
			ChildKeys:   foreignKey.ChildKeys,
			ParentKeys:  foreignKey.ParentKeys,
			ParentTable: foreignKey.ParentTable,
			Actions:     convertedActions,
		}
	}

	return convForeignKeys
}

func convertExtensionsToEngine(ext []*txpb.Extensions) []*types.Extension {
	convExtensions := make([]*types.Extension, len(ext))
	for i, e := range ext {
		exts := make([]*types.ExtensionConfig, len(e.Initialization))
		for i, init := range e.Initialization {
			exts[i] = &types.ExtensionConfig{
				Key:   init.Argument,
				Value: init.Value,
			}
		}

		convExtensions[i] = &types.Extension{
			Name:           e.Name,
			Initialization: exts,
			Alias:          e.Alias,
		}
	}

	return convExtensions
}

func convertProceduresToEngine(proc []*txpb.Procedure) []*types.Procedure {
	convProcedures := make([]*types.Procedure, len(proc))
	for i, p := range proc {
		t := &types.Procedure{
			Name:        p.Name,
			Annotations: p.Annotations,
			Public:      p.Public,
			Parameters:  convertParametersToEngine(p.Parameters),
			Modifiers:   convertModifiersToEngine(p.Modifiers),
			Body:        p.Body,
		}

		if p.ReturnTypes != nil {
			t.Returns = &types.ProcedureReturn{
				IsTable: p.ReturnTypes.IsTable,
				Fields:  make([]*types.NamedType, len(p.ReturnTypes.Columns)),
			}
			for j, r := range p.ReturnTypes.Columns {
				t.Returns.Fields[j] = &types.NamedType{
					Name: r.Name,
					Type: convertDataTypeToEngine(r.Type),
				}
			}
		}

		convProcedures[i] = t
	}

	return convProcedures
}

func convertParametersToEngine(incoming []*txpb.TypedVariable) []*types.ProcedureParameter {
	convParams := make([]*types.ProcedureParameter, len(incoming))
	for i, param := range incoming {
		convParams[i] = &types.ProcedureParameter{
			Name: param.Name,
			Type: convertDataTypeToEngine(param.Type),
		}
	}

	return convParams
}

func ConvertSchemaToPB(schema *types.Schema) (*txpb.Schema, error) {
	return &txpb.Schema{
		Owner:      schema.Owner,
		Name:       schema.Name,
		Tables:     convertTablesFromEngine(schema.Tables),
		Actions:    convertActionsFromEngine(schema.Actions),
		Extensions: convertExtensionsFromEngine(schema.Extensions),
		Procedures: convertProceduresFromEngine(schema.Procedures),
	}, nil
}

func convertTablesFromEngine(tables []*types.Table) []*txpb.Table {
	convTables := make([]*txpb.Table, len(tables))
	for i, table := range tables {
		convTable := &txpb.Table{
			Name:        table.Name,
			Columns:     convertColumnsFromEngine(table.Columns),
			Indexes:     convertIndexesFromEngine(table.Indexes),
			ForeignKeys: convertForeignKeysFromEngine(table.ForeignKeys),
		}
		convTables[i] = convTable
	}

	return convTables
}

func convertColumnsFromEngine(columns []*types.Column) []*txpb.Column {
	convColumns := make([]*txpb.Column, len(columns))
	for i, column := range columns {
		convColumn := &txpb.Column{
			Name:       column.Name,
			Type:       convertDataTypeFromEngine(column.Type),
			Attributes: convertAttributesFromEngine(column.Attributes),
		}
		convColumns[i] = convColumn
	}

	return convColumns
}

func convertDataTypeFromEngine(dataType *types.DataType) *txpb.DataType {
	return &txpb.DataType{
		Name:    dataType.Name,
		IsArray: dataType.IsArray,
	}
}

func convertAttributesFromEngine(attributes []*types.Attribute) []*txpb.Attribute {
	convAttributes := make([]*txpb.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttribute := &txpb.Attribute{
			Type:  attribute.Type.String(),
			Value: fmt.Sprint(attribute.Value),
		}
		convAttributes[i] = convAttribute
	}

	return convAttributes
}

func convertIndexesFromEngine(indexes []*types.Index) []*txpb.Index {
	convIndexes := make([]*txpb.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = &txpb.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type.String(),
		}
	}

	return convIndexes
}

func convertActionsFromEngine(actions []*types.Action) []*txpb.Action {
	convActions := make([]*txpb.Action, len(actions))
	for i, action := range actions {
		convActions[i] = &txpb.Action{
			Name:        action.Name,
			Public:      action.Public,
			Parameters:  action.Parameters,
			Modifiers:   convertModifiersFromEngine(action.Modifiers),
			Annotations: action.Annotations,
			Body:        action.Body,
		}
	}

	return convActions
}

func convertModifiersFromEngine(mods []types.Modifier) []string {
	convModifiers := make([]string, len(mods))
	for i, mod := range mods {
		convModifiers[i] = mod.String()
	}

	return convModifiers
}

func convertForeignKeysFromEngine(foreignKeys []*types.ForeignKey) []*txpb.ForeignKey {
	convForeignKeys := make([]*txpb.ForeignKey, len(foreignKeys))
	for i, foreignKey := range foreignKeys {
		convertedActions := make([]*txpb.ForeignKeyAction, len(foreignKey.Actions))

		for j, action := range foreignKey.Actions {
			convertedActions[j] = &txpb.ForeignKeyAction{
				On: action.On.String(),
				Do: action.Do.String(),
			}
		}

		convForeignKeys[i] = &txpb.ForeignKey{
			ChildKeys:   foreignKey.ChildKeys,
			ParentKeys:  foreignKey.ParentKeys,
			ParentTable: foreignKey.ParentTable,
			Actions:     convertedActions,
		}
	}

	return convForeignKeys
}

func convertExtensionsFromEngine(ext []*types.Extension) []*txpb.Extensions {
	convExtensions := make([]*txpb.Extensions, len(ext))
	for i, e := range ext {
		exts := make([]*txpb.Extensions_ExtensionConfig, len(e.Initialization))
		for i, init := range e.Initialization {
			exts[i] = &txpb.Extensions_ExtensionConfig{
				Argument: init.Key,
				Value:    init.Value,
			}
		}

		convExtensions[i] = &txpb.Extensions{
			Name:           e.Name,
			Initialization: exts,
			Alias:          e.Alias,
		}
	}

	return convExtensions
}

func convertProceduresFromEngine(proc []*types.Procedure) []*txpb.Procedure {
	convProcedures := make([]*txpb.Procedure, len(proc))
	for i, p := range proc {
		t := &txpb.Procedure{
			Name:        p.Name,
			Annotations: p.Annotations,
			Public:      p.Public,
			Parameters:  convertParameters(p.Parameters),
			Modifiers:   convertModifiersFromEngine(p.Modifiers),
			Body:        p.Body,
		}

		if p.Returns != nil {
			t.ReturnTypes = &txpb.ProcedureReturn{
				IsTable: p.Returns.IsTable,
				Columns: make([]*txpb.TypedVariable, len(p.Returns.Fields)),
			}
			for j, r := range p.Returns.Fields {
				t.ReturnTypes.Columns[j] = &txpb.TypedVariable{
					Name: r.Name,
					Type: convertDataTypeFromEngine(r.Type),
				}
			}
		}

		convProcedures[i] = t
	}

	return convProcedures
}

func convertParameters(incoming []*types.ProcedureParameter) []*txpb.TypedVariable {
	convParams := make([]*txpb.TypedVariable, len(incoming))
	for i, param := range incoming {
		convParams[i] = &txpb.TypedVariable{
			Name: param.Name,
			Type: convertDataTypeFromEngine(param.Type),
		}
	}

	return convParams
}
