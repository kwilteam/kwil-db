package dbretriever

import (
	"context"
	"fmt"
	"kwil/pkg/execution"
	"kwil/pkg/types/data_types"
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

func (q *dbRetriever) GetColumns(ctx context.Context, tableID int32) ([]*databases2.Column[anytype.KwilAny], error) {
	cols, err := q.gen.GetColumns(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf(`error getting columns for table %d: %w`, tableID, err)
	}

	columns := make([]*databases2.Column[anytype.KwilAny], len(cols))
	for i, col := range cols {
		attributes, err := q.GetAttributes(ctx, col.ID)
		if err != nil {
			return nil, fmt.Errorf(`error getting attributes for column %s: %w`, col.ColumnName, err)
		}

		columns[i] = &databases2.Column[anytype.KwilAny]{
			Name:       col.ColumnName,
			Type:       datatypes.DataType(col.ColumnType),
			Attributes: attributes,
		}
	}

	return columns, nil
}

func (q *dbRetriever) GetAttributes(ctx context.Context, columnID int32) ([]*databases2.Attribute[anytype.KwilAny], error) {
	attrs, err := q.gen.GetAttributes(ctx, columnID)
	if err != nil {
		return nil, fmt.Errorf(`error getting attributes for column %d: %w`, columnID, err)
	}

	attributes := make([]*databases2.Attribute[anytype.KwilAny], len(attrs))
	for i, attr := range attrs {
		value, err := anytype.NewFromSerial(attr.AttributeValue)
		if err != nil {
			return nil, fmt.Errorf(`error getting value for attribute %d: %w`, attr.AttributeType, err)
		}

		attributes[i] = &databases2.Attribute[anytype.KwilAny]{
			Type:  execution.AttributeType(attr.AttributeType),
			Value: value,
		}
	}

	return attributes, nil
}
