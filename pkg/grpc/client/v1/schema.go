package client

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
)

func (c *Client) GetSchema(ctx context.Context, dbid string) (*entity.Schema, error) {
	res, err := c.txClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	return convertSchema(res.Dataset), nil
}

func convertSchema(dataset *txpb.Dataset) *entity.Schema {
	return &entity.Schema{
		Owner:   dataset.Owner,
		Name:    dataset.Name,
		Tables:  convertTables(dataset.Tables),
		Actions: convertActions(dataset.Actions),
	}
}

func convertTables(tables []*txpb.Table) []*entity.Table {
	convTables := make([]*entity.Table, len(tables))
	for i, table := range tables {
		convTables[i] = &entity.Table{
			Name:    table.Name,
			Columns: convertColumns(table.Columns),
			Indexes: convertIndexes(table.Indexes),
		}
	}

	return convTables
}

func convertColumns(columns []*txpb.Column) []*entity.Column {
	convColumns := make([]*entity.Column, len(columns))
	for i, column := range columns {
		convColumns[i] = &entity.Column{
			Name:       column.Name,
			Type:       column.Type,
			Attributes: convertAttributes(column.Attributes),
		}
	}

	return convColumns
}

func convertAttributes(attributes []*txpb.Attribute) []*entity.Attribute {
	convAttributes := make([]*entity.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttributes[i] = &entity.Attribute{
			Type:  attribute.Type,
			Value: attribute.Value,
		}
	}

	return convAttributes
}

func convertIndexes(indexes []*txpb.Index) []*entity.Index {
	convIndexes := make([]*entity.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = &entity.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type,
		}
	}

	return convIndexes
}

func convertActions(actions []*txpb.Action) []*entity.Action {
	convActions := make([]*entity.Action, len(actions))
	for i, action := range actions {
		convActions[i] = &entity.Action{
			Name:       action.Name,
			Public:     action.Public,
			Mutability: action.Mutability,
			Inputs:     action.Inputs,
			Statements: action.Statements,
		}
	}

	return convActions
}
