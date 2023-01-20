package validation_test

import (
	"kwil/x/execution"
	"kwil/x/execution/validation"
	datatypes "kwil/x/types/data_types"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	"testing"
)

func testTables(t *testing.T, db *databases.Database[anytype.KwilAny]) {
	oldTables := mustCopy(db.Tables)
	// testing tables
	// test no tables
	db.Tables = []*databases.Table[anytype.KwilAny]{}
	err := validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for no tables")
	}
	db.Tables = mustCopy(oldTables)

	// test table with no name
	db.Tables[0].Name = ""
	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for no table name")
	}
	db.Tables = mustCopy(oldTables)

	// test table with no columns
	db.Tables[0].Columns = []*databases.Column[anytype.KwilAny]{}
	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for no table columns")
	}
	db.Tables = mustCopy(oldTables)

	// test column with invalid type
	db.Tables[0].Columns[0].Type = datatypes.DataType(-1)
	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for invalid column type")
	}
	db.Tables = mustCopy(oldTables)

	// invalid attribute type
	db.Tables[0].Columns[0].Attributes[0].Type = execution.AttributeType(-1)
	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for invalid attribute type")
	}
	db.Tables = mustCopy(oldTables)

	// try adding a boolean to a min attribute
	db.Tables[0].Columns[0].Attributes[0].Type = execution.MIN
	db.Tables[0].Columns[0].Attributes[0].Value, err = anytype.New(true)
	if err != nil {
		t.Errorf("failed to create new anytype: %v", err)
	}

	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for invalid attribute value")
	}
	db.Tables = mustCopy(oldTables)

	// try with min length
	db.Tables[0].Columns[0].Attributes[0].Type = execution.MIN
	db.Tables[0].Columns[0].Attributes[0].Value, err = anytype.New(1)
	if err != nil {
		t.Errorf("failed to create new anytype: %v", err)
	}
	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for invalid attribute value")
	}
	db.Tables = mustCopy(oldTables)
}
