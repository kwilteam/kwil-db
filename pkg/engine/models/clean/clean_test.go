package clean_test

import (
	"fmt"
	"kwil/pkg/engine/models/clean"
	"kwil/pkg/engine/models/mocks"
	"testing"
)

func Test_Clean(t *testing.T) {
	db := mocks.MOCK_DATASET1

	clean.Clean(&db)

	fmt.Println(db.Tables[0].Name)

}
