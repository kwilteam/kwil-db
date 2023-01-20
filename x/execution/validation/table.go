package validation

import (
	"fmt"
	"kwil/x/execution"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

func validateTables(d *databases.Database[anytype.KwilAny]) error {
	// check amount
	if len(d.Tables) > execution.MAX_TABLE_COUNT {
		return fmt.Errorf(`database must have at most %d tables`, execution.MAX_TABLE_COUNT)
	}

	// check unique table names and validate tables
	tables := make(map[string]struct{})
	for _, table := range d.Tables {
		// check if table name is unique
		if _, ok := tables[table.Name]; ok {
			return fmt.Errorf(`duplicate table name "%s"`, table.Name)
		}
		tables[table.Name] = struct{}{}

		err := ValidateTable(table)
		if err != nil {
			return fmt.Errorf(`error on table "%s": %w`, table.Name, err)
		}
	}
	return nil
}

func ValidateTable(table *databases.Table[anytype.KwilAny]) error {
	// check name and name length
	err := CheckName(table.Name, execution.MAX_TABLE_NAME_LENGTH)
	if err != nil {
		return err
	}

	// check amount of columns
	if len(table.Columns) == 0 {
		return fmt.Errorf(`table must have at least one column`)
	}
	if len(table.Columns) > execution.MAX_COLUMNS_PER_TABLE {
		return fmt.Errorf(`table must have at most %d columns`, execution.MAX_COLUMNS_PER_TABLE)
	}

	cols := make(map[string]struct{})
	for _, col := range table.Columns {
		// check if column name is unique
		if _, ok := cols[col.Name]; ok {
			return fmt.Errorf(`duplicate column name "%s"`, col.Name)
		}
		cols[col.Name] = struct{}{}

		err := ValidateColumn(col)
		if err != nil {
			return fmt.Errorf(`error on column "%s": %w`, col.Name, err)
		}
	}

	return nil
}
