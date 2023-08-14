package client

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	res, err := c.txClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	return convertSchema(res.Schema), nil
}

func convertSchema(dataset *txpb.Schema) *transactions.Schema {
	return &transactions.Schema{
		Owner:   dataset.Owner,
		Name:    dataset.Name,
		Tables:  convertTables(dataset.Tables),
		Actions: convertActions(dataset.Actions),
	}
}

func convertTables(tables []*txpb.Table) []*transactions.Table {
	convTables := make([]*transactions.Table, len(tables))
	for i, table := range tables {
		convTables[i] = &transactions.Table{
			Name:    table.Name,
			Columns: convertColumns(table.Columns),
			Indexes: convertIndexes(table.Indexes),
		}
	}

	return convTables
}

func convertColumns(columns []*txpb.Column) []*transactions.Column {
	convColumns := make([]*transactions.Column, len(columns))
	for i, column := range columns {
		convColumns[i] = &transactions.Column{
			Name:       column.Name,
			Type:       column.Type,
			Attributes: convertAttributes(column.Attributes),
		}
	}

	return convColumns
}

func convertAttributes(attributes []*txpb.Attribute) []*transactions.Attribute {
	convAttributes := make([]*transactions.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttributes[i] = &transactions.Attribute{
			Type:  attribute.Type,
			Value: attribute.Value,
		}
	}

	return convAttributes
}

func convertIndexes(indexes []*txpb.Index) []*transactions.Index {
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

func convertActions(actions []*txpb.Action) []*transactions.Action {
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
