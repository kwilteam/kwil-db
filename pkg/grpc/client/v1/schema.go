package client

import (
	"context"

	"github.com/kwilteam/kuneiform/schema"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) GetSchema(ctx context.Context, dbid string) (*schema.Schema, error) {
	res, err := c.txClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	return convertSchema(res.Dataset), nil
}

func convertSchema(dataset *txpb.Dataset) *schema.Schema {
	return &schema.Schema{
		Owner:   dataset.Owner,
		Name:    dataset.Name,
		Tables:  convertTables(dataset.Tables),
		Actions: convertActions(dataset.Actions),
	}
}

func convertTables(tables []*txpb.Table) []schema.Table {
	convTables := make([]schema.Table, len(tables))
	for i, table := range tables {
		convTables[i] = schema.Table{
			Name:    table.Name,
			Columns: convertColumns(table.Columns),
			Indexes: convertIndexes(table.Indexes),
		}
	}

	return convTables
}

func convertColumns(columns []*txpb.Column) []schema.Column {
	convColumns := make([]schema.Column, len(columns))
	for i, column := range columns {
		convColumns[i] = schema.Column{
			Name:       column.Name,
			Type:       schema.ColumnType(column.Type),
			Attributes: convertAttributes(column.Attributes),
		}
	}

	return convColumns
}

func convertAttributes(attributes []*txpb.Attribute) []schema.Attribute {
	convAttributes := make([]schema.Attribute, len(attributes))
	for i, attribute := range attributes {
		convAttributes[i] = schema.Attribute{
			Type:  schema.AttributeType(attribute.Type),
			Value: attribute.Value,
		}
	}

	return convAttributes
}

func convertIndexes(indexes []*txpb.Index) []schema.Index {
	convIndexes := make([]schema.Index, len(indexes))
	for i, index := range indexes {
		convIndexes[i] = schema.Index{
			Name:    index.Name,
			Columns: index.Columns,
			Type:    schema.IndexType(index.Type),
		}
	}

	return convIndexes
}

func convertActions(actions []*txpb.Action) []schema.Action {
	convActions := make([]schema.Action, len(actions))
	for i, action := range actions {
		convActions[i] = schema.Action{
			Name:       action.Name,
			Public:     action.Public,
			Inputs:     action.Inputs,
			Statements: action.Statements,
		}
	}

	return convActions
}
