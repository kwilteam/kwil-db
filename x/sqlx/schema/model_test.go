package schema

import (
	"fmt"
	"os"
	"testing"
)

func Test_ReadYaml(t *testing.T) {
	// read in test.yaml

	bts, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	db := &Database{}
	err = db.UnmarshalYAML(bts)
	if err != nil {
		t.Fatal(err)
	}

	if db.Owner != "kwil" {
		t.Fatal("owner should be kwil")
	}

	if db.Name != "mydb" {
		t.Fatal("name should be mydb")
	}

	err = db.Validate()
	if err != nil {
		t.Fatal(err)
	}

	q := db.Queries.Inserts["createUser"]
	if q == nil {
		t.Fatal("createUser is nil")
	}

	st, err := q.Prepare(db)
	if err != nil {
		t.Fatal(err)
	}

	inpts := make(UserInputs)
	inpts["first_name"] = "kwil"
	inpts["last_name"] = "wilson"
	inpts["age"] = "42"
	inpts["user_id"] = "1"

	_, err = st.PrepareInputs(&inpts)
	if err != nil {
		t.Fatal(err)
	}

	bts, err = st.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(bts))
	panic("")
}

func Test_UpdateStmt(t *testing.T) {
	// read in test.yaml

	bts, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	db := &Database{}
	err = db.UnmarshalYAML(bts)
	if err != nil {
		t.Fatal(err)
	}

	if db.Owner != "kwil" {
		t.Fatal("owner should be kwil")
	}

	if db.Name != "mydb" {
		t.Fatal("name should be mydb")
	}

	err = db.Validate()
	if err != nil {
		t.Fatal(err)
	}

	q := db.Queries.Updates["updateUser"]
	if q == nil {
		t.Fatal("updateUser is nil")
	}

	st, err := q.Prepare(db)
	if err != nil {
		t.Fatal(err)
	}

	inpts := make(UserInputs)
	inpts["first_name"] = "kwil"
	inpts["last_name"] = "wilson"
	inpts["age"] = "42"
	inpts["user_id"] = "1"

	ins, err := st.PrepareInputs(&inpts)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(st.Statement)
	fmt.Println(ins)
	fmt.Println(st.UserInputs)

	panic("	")
}

func Test_DeleteStmt(t *testing.T) {
	// read in test.yaml

	bts, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	db := &Database{}
	err = db.UnmarshalYAML(bts)
	if err != nil {
		t.Fatal(err)
	}

	if db.Owner != "kwil" {
		t.Fatal("owner should be kwil")
	}

	if db.Name != "mydb" {
		t.Fatal("name should be mydb")
	}

	err = db.Validate()
	if err != nil {
		t.Fatal(err)
	}

	q := db.Queries.Deletes["removeUser"]
	if q == nil {
		t.Fatal("deleteUser is nil")
	}

	st, err := q.Prepare(db)
	if err != nil {
		t.Fatal(err)
	}

	inpts := make(UserInputs)
	inpts["user_id"] = "1"

	ins, err := st.PrepareInputs(&inpts)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(st.Statement)
	fmt.Println(ins)

	panic("	")
}

func Test_DBMarshalling(t *testing.T) {
	bts, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	db := &Database{}
	err = db.UnmarshalYAML(bts)
	if err != nil {
		t.Fatal(err)
	}

	bts, err = db.EncodeGOB()
	if err != nil {
		t.Fatal(err)
	}

	db2 := &Database{}
	err = db2.DecodeGOB(bts)
	if err != nil {
		t.Fatal(err)
	}

	bts2, err := db2.EncodeGOB()
	if err != nil {
		t.Fatal(err)
	}

	if len(bts) != len(bts2) {
		t.Fatal("bytes are not equal")
	}

	panic("")
}
