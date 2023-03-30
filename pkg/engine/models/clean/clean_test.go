package clean_test

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/models/clean"
	"os"
	"testing"
)

func Test_Clean(t *testing.T) {
	db, err := dbFromJson("../mocks/test_clean.json")
	if err != nil {
		t.Fatal(err)
	}

	clean.Clean(db)

	fmt.Println(db.Tables[0].Name)

}

func dbFromJson(path string) (*models.Dataset, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	db := &models.Dataset{}
	err = json.Unmarshal(b, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
