package models_test

import (
	"encoding/json"
	"fmt"
	"kwil/x/sqlx/models"
	"os"
	"testing"
)

func Test_UnmarshalDatabase(t *testing.T) {

	bts, err := os.ReadFile("test.json")
	if err != nil {
		t.Fatal(err)
	}

	var db models.Database
	err = db.FromJSON(bts)
	if err != nil {
		t.Fatal(err)
	}

	db.Clean()

	err = db.Validate()
	if err != nil {
		t.Fatal(err)
	}

	ddl, err := db.GenerateDDL()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(ddl)

	execs, err := db.PrepareQueries()
	if err != nil {
		t.Fatal(err)
	}

	for _, exec := range execs {
		fmt.Println(exec.Statement)
	}

	panic("")

}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}
