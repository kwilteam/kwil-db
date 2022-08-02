package txsvc

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
)

func (s *Service) GetSchema(ctx context.Context, req *txpb.GetSchemaRequest) (*txpb.GetSchemaResponse, error) {
	schema, err := s.executor.GetSchema(ctx, req.Dbid)
	if err != nil {
		return nil, err
	}

	return &txpb.GetSchemaResponse{
		Dataset: convertSchema(schema),
	}, nil
}

func convertSchema(schema *entity.Schema) *txpb.Dataset {
	return &txpb.Dataset{
		Owner:   schema.Owner,
		Name:    schema.Name,
		Tables:  convertTables(schema.Tables),
		Actions: convertActions(schema.Actions),
	}
}

func convertTables(tables []*entity.Table) []*txpb.Table {
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

func convertColumns(columns []*entity.Column) []*txpb.Column {
	convColumns := make([]*txpb.Column, len(columns))
	for i, column := range columns {
		convColumn := &txpb.Column{
			Name:       column.Name,
			Type:       column.Type,
			Attributes: convertAttributes(column.Attributes),
		}
		convColumns[i] = convColumn
	}

	return convColumns
}

func convertAttributes(attributes []*entity.Attribute) []*txpb.Attribute {
	convAttributes := make([]*txpb.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttribute := &txpb.Attribute{
			Type:  attribute.Type,
			Value: fmt.Sprint(attribute.Value),
		}
		convAttributes[i] = convAttribute
	}

	return convAttributes
}

func convertIndexes(indexes []*entity.Index) []*txpb.Index {
	convIndexes := make([]*txpb.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = &txpb.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type,
		}
	}

	return convIndexes
}

func convertActions(actions []*entity.Action) []*txpb.Action {
	convActions := make([]*txpb.Action, len(actions))
	for i, action := range actions {
		convActions[i] = &txpb.Action{
			Name:       action.Name,
			Public:     action.Public,
			Inputs:     action.Inputs,
			Statements: action.Statements,
		}
	}

	return convActions
}
