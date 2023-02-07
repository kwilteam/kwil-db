package dbretriever

import (
	"context"
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/types/data_types"
	"kwil/pkg/types/data_types/any_type"
)

func (q *dbRetriever) GetColumns(ctx context.Context, tableID int32) ([]*databases.Column[anytype.KwilAny], error) {
	cols, err := q.gen.GetColumns(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf(`error getting columns for table %d: %w`, tableID, err)
	}

	columns := make([]*databases.Column[anytype.KwilAny], len(cols))
	for i, col := range cols {
		attributes, err := q.GetAttributes(ctx, col.ID)
		if err != nil {
			return nil, fmt.Errorf(`error getting attributes for column %s: %w`, col.ColumnName, err)
		}

		columns[i] = &databases.Column[anytype.KwilAny]{
			Name:       col.ColumnName,
			Type:       datatypes.DataType(col.ColumnType),
			Attributes: attributes,
		}
	}

	return columns, nil
}

func (q *dbRetriever) GetAttributes(ctx context.Context, columnID int32) ([]*databases.Attribute[anytype.KwilAny], error) {
	attrs, err := q.gen.GetAttributes(ctx, columnID)
	if err != nil {
		return nil, fmt.Errorf(`error getting attributes for column %d: %w`, columnID, err)
	}

	attributes := make([]*databases.Attribute[anytype.KwilAny], len(attrs))
	for i, attr := range attrs {
		value, err := anytype.NewFromSerial(attr.AttributeValue)
		if err != nil {
			return nil, fmt.Errorf(`error getting value for attribute %d: %w`, attr.AttributeType, err)
		}

		attributes[i] = &databases.Attribute[anytype.KwilAny]{
			Type:  databases.AttributeType(attr.AttributeType),
			Value: value,
		}
	}

	return attributes, nil
}
