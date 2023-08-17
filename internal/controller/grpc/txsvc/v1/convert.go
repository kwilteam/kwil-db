package txsvc

import (
	"fmt"

	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func convertTransaction(incoming *txpb.Transaction) (*transactions.Transaction, error) {
	payloadType := transactions.PayloadType(incoming.Body.PayloadType)
	if !payloadType.Valid() {
		return nil, fmt.Errorf("invalid payload type: %s", incoming.Body.PayloadType)
	}

	if incoming.Signature == nil {
		return nil, fmt.Errorf("transaction signature cannot be nil")
	}

	convSignature, err := convertSignature(incoming.Signature)
	if err != nil {
		return nil, err
	}

	sender, err := crypto.PublicKeyFromBytes(convSignature.Type.KeyType(), incoming.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender public key: %s", err.Error())
	}

	bigFee, ok := big.NewInt(0).SetString(incoming.Body.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("invalid fee: %s", incoming.Body.Fee)
	}

	return &transactions.Transaction{
		Signature: convSignature,
		Body: &transactions.TransactionBody{
			PayloadType: payloadType,
			Payload:     incoming.Body.Payload,
			Nonce:       incoming.Body.Nonce,
			Fee:         bigFee,
			Salt:        incoming.Body.Salt,
		},
		Sender: sender,
	}, nil
}

func newEmptySignature() (bytes []byte, sigType crypto.SignatureType) {
	return []byte{}, crypto.SIGNATURE_TYPE_EMPTY
}

func convertSignature(sig *txpb.Signature) (*crypto.Signature, error) {
	if sig == nil {
		emptyBts, emptyType := newEmptySignature()
		return &crypto.Signature{
			Signature: emptyBts,
			Type:      emptyType,
		}, nil
	}

	sigType := crypto.SignatureLookUp(sig.SignatureType)
	if err := sigType.IsValid(); err != nil {
		return nil, err
	}

	return &crypto.Signature{
		Signature: sig.SignatureBytes,
		Type:      sigType,
	}, nil
}

func convertSchemaFromEngine(schema *engineTypes.Schema) (*txpb.Schema, error) {
	actions, err := convertActionsFromEngine(schema.Procedures)
	if err != nil {
		return nil, err
	}
	return &txpb.Schema{
		Owner:   schema.Owner,
		Name:    schema.Name,
		Tables:  convertTablesFromEngine(schema.Tables),
		Actions: actions,
	}, nil
}

func convertTablesFromEngine(tables []*engineTypes.Table) []*txpb.Table {
	convTables := make([]*txpb.Table, len(tables))
	for i, table := range tables {
		convTable := &txpb.Table{
			Name:    table.Name,
			Columns: convertColumnsFromEngine(table.Columns),
			Indexes: convertIndexesFromEngine(table.Indexes),
		}
		convTables[i] = convTable
	}

	return convTables
}

func convertColumnsFromEngine(columns []*engineTypes.Column) []*txpb.Column {
	convColumns := make([]*txpb.Column, len(columns))
	for i, column := range columns {
		convColumn := &txpb.Column{
			Name:       column.Name,
			Type:       column.Type.String(),
			Attributes: convertAttributesFromEngine(column.Attributes),
		}
		convColumns[i] = convColumn
	}

	return convColumns
}

func convertAttributesFromEngine(attributes []*engineTypes.Attribute) []*txpb.Attribute {
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

func convertIndexesFromEngine(indexes []*engineTypes.Index) []*txpb.Index {
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

func convertActionsFromEngine(actions []*engineTypes.Procedure) ([]*txpb.Action, error) {

	convActions := make([]*txpb.Action, len(actions))
	for i, action := range actions {
		mutability, auxiliaries, err := convertModifiersFromEngine(action.Modifiers)
		if err != nil {
			return nil, err
		}

		convActions[i] = &txpb.Action{
			Name:        action.Name,
			Public:      action.Public,
			Mutability:  mutability,
			Auxiliaries: auxiliaries,
			Inputs:      action.Args,
			Statements:  action.Statements,
		}
	}

	return convActions, nil
}

func convertModifiersFromEngine(mods []engineTypes.Modifier) (mutability string, auxiliaries []string, err error) {
	auxiliaries = make([]string, 0)
	mutability = "UPDATE"
	for _, mod := range mods {
		switch mod {
		case engineTypes.ModifierAuthenticated:
			auxiliaries = append(auxiliaries, "AUTHENTICATED")
		case engineTypes.ModifierView:
			mutability = "VIEW"
		case engineTypes.ModifierOwner:
			auxiliaries = append(auxiliaries, "OWNER")
		default:
			return "", nil, fmt.Errorf("unknown modifier type: %v", mod)
		}
	}

	return mutability, auxiliaries, nil
}
