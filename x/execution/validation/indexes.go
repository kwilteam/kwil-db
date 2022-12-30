package validation

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/execution/dto"
	"sort"
)

func validateIndexes(d *dto.Database) error {
	// check amount of indexes
	if len(d.Indexes) > execution.MAX_INDEX_COUNT {
		return fmt.Errorf(`database must have at most %d indexes`, execution.MAX_INDEX_COUNT)
	}

	// check unique index names and validate indexes
	indexes := make(map[string]struct{})
	// indexes must also check unique columns and type
	indexColAndType := make(map[string]struct{})
	for _, index := range d.Indexes {
		// check if index name is unique
		if _, ok := indexes[index.Name]; ok {
			return fmt.Errorf(`duplicate index name "%s"`, index.Name)
		}
		indexes[index.Name] = struct{}{}

		// check if index columns and type are unique
		// first sort columns
		cols := make([]string, len(index.Columns))
		copy(cols, index.Columns)
		sort.Strings(cols)
		// then create key
		key := fmt.Sprintf("%s:%s", index.Using.String(), cols)
		if _, ok := indexColAndType[key]; ok {
			return fmt.Errorf(`duplicate index columns and type.  Columns: "%s".  Type: "%s"`, index.Columns, index.Using.String())
		}
		indexColAndType[key] = struct{}{}

		err := ValidateIndex(index, d)
		if err != nil {
			return fmt.Errorf(`error on index "%s": %w`, index.Name, err)
		}
	}

	return nil
}

func ValidateIndex(index *dto.Index, db *dto.Database) error {
	// check if index name is valid
	err := CheckName(index.Name, execution.MAX_INDEX_NAME_LENGTH)
	if err != nil {
		return err
	}

	// check if index type is valid
	if !index.Using.IsValid() {
		return fmt.Errorf(`unknown index type: %d`, index.Using.Int())
	}

	// check if index table is valid
	table := db.GetTable(index.Table)
	if table == nil {
		return fmt.Errorf(`table "%s" does not exist`, index.Table)
	}

	// check if index columns are valid
	columns := make(map[string]struct{})
	for _, col := range index.Columns {
		// check if column is unique
		if _, ok := columns[col]; ok {
			return fmt.Errorf(`duplicate column "%s"`, col)
		}
		columns[col] = struct{}{}

		// check if column exists
		if table.GetColumn(col) == nil {
			return fmt.Errorf(`column "%s" does not exist`, col)
		}
	}

	return nil
}
