package client

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/serialize"
)

func (c *Client) GetSchema(ctx context.Context, dbid string) (*serialize.Schema, error) {
	res, err := c.txClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	return convertSchema(res.Dataset), nil
}

func convertSchema(dataset *txpb.Dataset) *serialize.Schema {
	return &serialize.Schema{
		Owner:   dataset.Owner,
		Name:    dataset.Name,
		Tables:  convertTables(dataset.Tables),
		Actions: convertActions(dataset.Actions),
	}
}

func convertTables(tables []*txpb.Table) []*serialize.Table {
	convTables := make([]*serialize.Table, len(tables))
	for i, table := range tables {
		convTables[i] = &serialize.Table{
			Name:    table.Name,
			Columns: convertColumns(table.Columns),
			Indexes: convertIndexes(table.Indexes),
		}
	}

	return convTables
}

func convertColumns(columns []*txpb.Column) []*serialize.Column {
	convColumns := make([]*serialize.Column, len(columns))
	for i, column := range columns {
		convColumns[i] = &serialize.Column{
			Name:       column.Name,
			Type:       column.Type,
			Attributes: convertAttributes(column.Attributes),
		}
	}

	return convColumns
}

func convertAttributes(attributes []*txpb.Attribute) []*serialize.Attribute {
	convAttributes := make([]*serialize.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttributes[i] = &serialize.Attribute{
			Type:  attribute.Type,
			Value: attribute.Value,
		}
	}

	return convAttributes
}

func convertIndexes(indexes []*txpb.Index) []*serialize.Index {
	convIndexes := make([]*serialize.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = &serialize.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    index.Type,
		}
	}

	return convIndexes
}

func convertActions(actions []*txpb.Action) []*serialize.Action {
	convActions := make([]*serialize.Action, len(actions))
	for i, action := range actions {
		convActions[i] = &serialize.Action{
			Name:       action.Name,
			Public:     action.Public,
			Mutability: action.Mutability,
			Inputs:     action.Inputs,
			Statements: action.Statements,
		}
	}

	return convActions
}
