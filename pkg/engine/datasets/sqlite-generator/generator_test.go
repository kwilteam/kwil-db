package sqlitegenerator_test

import (
	"fmt"
	sqlitegenerator "kwil/pkg/engine/datasets/sqlite-generator"
	"kwil/pkg/engine/models/mocks"
	"testing"
)

func Test_Generate(t *testing.T) {
	ddl, err := sqlitegenerator.GenerateDDL(&mocks.MOCK_TABLE1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ddl)

	ddl2, err := sqlitegenerator.GenerateDDL(&mocks.MOCK_TABLE2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ddl2)
}
