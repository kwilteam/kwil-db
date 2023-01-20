package validation_test

import (
	"kwil/x/execution/mocks"
	"kwil/x/execution/validation"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	"testing"

	"github.com/mitchellh/copystructure"
)

func Test_Validation(t *testing.T) {

	// happy day case
	err := validation.ValidateDatabase(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to validate database: %v", err)
	}

	copyDB := mustCopy(&mocks.Db1)

	// we can now modify the copy and test incorrect cases
	testNames(t, copyDB)

	// database with no tables
	testTables(t, copyDB)

}

func testNames(t *testing.T, db *databases.Database[anytype.KwilAny]) {
	// testing names.  The same name check is used for all names, so we only need to test one
	// test no name
	db.Name = ""
	err := validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for no name")
	}

	// name too long
	db.Name = "abc1234567891011121314151617181920212223242526272829303132333435363738394041424344454647484950"
	err = validation.ValidateDatabase(db)
	if err == nil {
		t.Errorf("expected error for name too long")
	}
	db.Name = "abc"
}

func mustCopy[T any](i T) T {
	c, err := copystructure.Copy(i)
	if err != nil {
		panic(err)
	}
	return c.(T)
}
