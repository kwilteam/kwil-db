package dbretriever

import (
	"context"
	"fmt"
	"kwil/x/execution"
	"kwil/x/types/databases"
	"kwil/x/utils/serialize"
)

func (q *dbRetriever) GetColumns(ctx context.Context, tableID int32) ([]*databases.Column, error) {
	cols, err := q.gen.GetColumns(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf(`error getting columns for table %d: %w`, tableID, err)
	}

	columns := make([]*databases.Column, len(cols))
	for i, col := range cols {
		attributes, err := q.GetAttributes(ctx, col.ID)
		if err != nil {
			return nil, fmt.Errorf(`error getting attributes for column %s: %w`, col.ColumnName, err)
		}

		columns[i] = &databases.Column{
			Name:       col.ColumnName,
			Type:       execution.DataType(col.ColumnType),
			Attributes: attributes,
		}
	}

	return columns, nil
}

func (q *dbRetriever) GetAttributes(ctx context.Context, columnID int32) ([]*databases.Attribute, error) {
	attrs, err := q.gen.GetAttributes(ctx, columnID)
	if err != nil {
		return nil, fmt.Errorf(`error getting attributes for column %d: %w`, columnID, err)
	}

	attributes := make([]*databases.Attribute, len(attrs))
	for i, attr := range attrs {
		val, err := serialize.UnmarshalType(attr.AttributeValue)
		if err != nil {
			return nil, fmt.Errorf(`error unmarshaling attribute value: %w`, err)
		}

		attributes[i] = &databases.Attribute{
			Type:  execution.AttributeType(attr.AttributeType),
			Value: val,
		}
	}

	return attributes, nil
}
