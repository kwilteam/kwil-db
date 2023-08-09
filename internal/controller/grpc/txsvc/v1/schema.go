package txsvc

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	schema, err := s.engine.GetSchema(ctx, req.Dbid)
	if err != nil {
		return nil, err
	}

	txSchema, err := convertSchema(schema)
	if err != nil {
		return nil, err
	}

	return &txpb.GetSchemaResponse{
		Dataset: txSchema,
	}, nil
}

func convertSchema(schema *engineTypes.Schema) (*txpb.Dataset, error) {
	actions, err := convertActions(schema.Procedures)
	if err != nil {
		return nil, err
	}
	return &txpb.Dataset{
		Owner:   schema.Owner,
		Name:    schema.Name,
		Tables:  convertTables(schema.Tables),
		Actions: actions,
	}, nil
}

func convertTables(tables []*engineTypes.Table) []*txpb.Table {
	convTables := make([]*txpb.Table, len(tables))
	for i, table := range tables {
		convTable := &txpb.Table{
			Name:    table.Name,
			Columns: convertColumns(table.Columns),
			Indexes: convertIndexes(table.Indexes),
		}
		convTables[i] = convTable
	}

	return convTables
}

func convertColumns(columns []*engineTypes.Column) []*txpb.Column {
	convColumns := make([]*txpb.Column, len(columns))
	for i, column := range columns {
		convColumn := &txpb.Column{
			Name:       column.Name,
			Type:       column.Type.String(),
			Attributes: convertAttributes(column.Attributes),
		}
		convColumns[i] = convColumn
	}

	return convColumns
}

func convertAttributes(attributes []*engineTypes.Attribute) []*txpb.Attribute {
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

func convertIndexes(indexes []*engineTypes.Index) []*txpb.Index {
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

func convertActions(actions []*engineTypes.Procedure) ([]*txpb.Action, error) {

	convActions := make([]*txpb.Action, len(actions))
	for i, action := range actions {
		mutability, auxiliaries, err := convertModifiers(action.Modifiers)
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

func convertModifiers(mods []engineTypes.Modifier) (mutability string, auxiliaries []string, err error) {
	auxiliaries = make([]string, 0)
	mutability = "UPDATE"
	for _, mod := range mods {
		switch mod {
		case engineTypes.ModifierAuthenticated:
			auxiliaries = append(auxiliaries, "AUTHENTICATED")
		case engineTypes.ModifierView:
			mutability = "VIEW"
		// TODO: add modifier owner once merged
		default:
			return "", nil, fmt.Errorf("unknown modifier type: %v", mod)
		}
	}

	return mutability, auxiliaries, nil
}
