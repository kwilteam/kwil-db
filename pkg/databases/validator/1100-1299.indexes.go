package validator

import (
	"fmt"
	"kwil/pkg/databases"
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
	for _, index := range v.DB.Indexes {
		// check if index name is unique
		if _, ok := indexNames[index.Name]; ok {
			return violation(errorCode1100, fmt.Errorf(`duplicate index name "%s"`, index.Name))
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
			return violation(errorCode1102, fmt.Errorf(`duplicate index columns and type.  Columns: "%s".  Type: "%s"`, index.Columns, index.Using.String()))
		}
		indexColAndType[key] = struct{}{}

		err := v.ValidateIndex(index)
		if err != nil {
			return fmt.Errorf(`error on index "%s": %w`, index.Name, err)
		}
	}

	return nil
}

func (v *Validator) validateIndexCount() error {
	if len(v.DB.Indexes) > MAX_INDEX_COUNT {
		return violation(errorCode1101, fmt.Errorf(`database has too many indexes: %v > %v`, len(v.DB.Indexes), MAX_INDEX_COUNT))
	}

	return nil
}

func (v *Validator) ValidateIndex(index *databases.Index) error {
	// check if index name is valid
	err := CheckName(index.Name, MAX_INDEX_NAME_LENGTH)
	if err != nil {
		return violation(errorCode1200, fmt.Errorf(`invalid index name "%s": %w`, index.Name, err))
	}

	if isReservedWord(index.Name) {
		return violation(errorCode1201, fmt.Errorf(`index name "%s" is a reserved word`, index.Name))
	}

	// check if index type is valid
	if !index.Using.IsValid() {
		return violation(errorCode1202, fmt.Errorf(`invalid index type "%d"`, index.Using.Int()))
	}

	// check if index table is valid
	table := v.DB.GetTable(index.Table)
	if table == nil {
		return violation(errorCode1203, fmt.Errorf(`table "%s" does not exist`, index.Table))
	}

	// check that index columns are not empty
	if len(index.Columns) == 0 {
		return violation(errorCode1206, fmt.Errorf(`index has no columns`))
	}

	// check that there aren't too many columns
	if len(index.Columns) > MAX_INDEX_COLUMNS {
		return violation(errorCode1207, fmt.Errorf(`index has too many columns: %v > %v`, len(index.Columns), MAX_INDEX_COLUMNS))
	}

	// check if index columns are valid
	columns := make(map[string]struct{})
	for _, col := range index.Columns {
		// check if column is unique
		if _, ok := columns[col]; ok {
			return violation(errorCode1205, fmt.Errorf(`duplicate column "%s"`, col))
		}
		columns[col] = struct{}{}

		// check if column exists
		if table.GetColumn(col) == nil {
			return violation(errorCode1204, fmt.Errorf(`column "%s" does not exist`, col))
		}
	}

	return nil
}
