package validator

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/types/databases"
	"sort"
)

func (v *Validator) validateIndexes() error {
	// check if there are too many indexes
	err := v.validateIndexCount()
	if err != nil {
		return fmt.Errorf(`invalid index count: %w`, err)
	}

	// check unique index names and validate indexes
	indexNames := make(map[string]struct{})
	// indexes must also check unique columns and type
	indexColAndType := make(map[string]struct{})
	for _, index := range v.db.Indexes {
		// check if index name is unique
		if _, ok := indexNames[index.Name]; ok {
			return fmt.Errorf(`duplicate index name "%s"`, index.Name)
		}
		indexNames[index.Name] = struct{}{}

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

		err := v.validateIndex(index)
		if err != nil {
			return fmt.Errorf(`error on index "%s": %w`, index.Name, err)
		}
	}

	return nil
}

func (v *Validator) validateIndexCount() error {
	if len(v.db.Indexes) > execution.MAX_INDEX_COUNT {
		return fmt.Errorf(`too many indexes: %v > %v`, len(v.db.Indexes), execution.MAX_INDEX_COUNT)
	}

	return nil
}

func (v *Validator) validateIndex(index *databases.Index) error {
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
	table := v.db.GetTable(index.Table)
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
