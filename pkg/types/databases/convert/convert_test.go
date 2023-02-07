package convert_test

import (
	"kwil/pkg/types/databases/convert"
	"kwil/pkg/types/databases/mocks"
	"testing"
)

func Test_Convert(t *testing.T) {
	db := mocks.Db1

	db2, err := convert.KwilAny.DatabaseToBytes(&db)
	if err != nil {
		t.Errorf("error converting database: %v", err)
	}

	if db2 == nil {
		t.Errorf("database is nil")
	}

	db3, err := convert.Bytes.DatabaseToKwilAny(db2)
	if err != nil {
		t.Errorf("error converting database: %v", err)
	}
	if db3 == nil {
		t.Errorf("database is nil")
	}

	if db3 == nil {
		t.Errorf("database is nil")
	}

	//lint:ignore SA5011 false positive
	for i, tbl := range db3.Tables {
		if tbl.Name != db.Tables[i].Name {
			t.Errorf("table name mismatch: %v != %v", tbl.Name, db.Tables[i].Name)
		}
	}
}
