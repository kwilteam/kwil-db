package schema

import (
	"os"
	"testing"
)

func Test_ReadYaml(t *testing.T) {
	// read in test.yaml

	bts, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	db, err := ReadYaml(bts)
	if err != nil {
		t.Fatal(err)
	}

	if db.Owner != "kwil" {
		t.Fatal("owner should be kwil")
	}

	if db.Name != "mydb" {
		t.Fatal("name should be mydb")
	}

	/*err = Verify(db)
	if err != nil {
		t.Fatal(err)
	}*/
	err = db.Validate()
	if err != nil {
		t.Fatal(err)
	}

}
