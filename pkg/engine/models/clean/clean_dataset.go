package clean

import (
	"fmt"
	"kwil/pkg/engine/models"
)

// CleanDataset cleans a dataset and coerces attribute values
func CleanDataset(dataset *models.Dataset) error {
	Clean(dataset)

	for _, table := range dataset.Tables {
		for _, col := range table.Columns {
			err := coerceAttributeValues(col)
			if err != nil {
				return fmt.Errorf("error coercing attribute values: %w", err)
			}
		}
	}

	return nil
}

// coerceAttributeValues coerces the attribute values to the correct type
func coerceAttributeValues(col *models.Column) error {
	for _, attr := range col.Attributes {
		err := attr.Coerce(col.Type)
		if err != nil {
			return fmt.Errorf("error coercing attribute value: %w", err)
		}
	}
	return nil
}
