package clean_test

import (
	"bytes"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models/clean"
	"github.com/kwilteam/kwil-db/pkg/engine/models/mocks"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"testing"
)

func Test_Clean(t *testing.T) {
	db := mocks.MOCK_DATASET1

	err := clean.CleanDataset(&db)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(db.Tables[0].Name)

	// try to set incorrect attribute value that should be coerced
	// table 0, column 1, attribute 1 has a max length of 100
	db2 := mocks.MOCK_DATASET1
	db2.Tables[0].Columns[1].Attributes[1].Value = types.NewMust("101").Bytes()

	// ensure it is a string
	if !bytes.Equal(db2.Tables[0].Columns[1].Attributes[1].Value, types.NewMust("101").Bytes()) {
		t.Fatal(`expected 101 as a string, got: `, string(db2.Tables[0].Columns[1].Attributes[1].Value))
	}

	err = clean.CleanDataset(&db2)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(db2.Tables[0].Columns[1].Attributes[1].Value, types.NewMust(101).Bytes()) {
		t.Fatal(`expected 101 as an int serialized, got: `, string(db2.Tables[0].Columns[1].Attributes[1].Value))
	}
}
