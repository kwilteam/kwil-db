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

	db, err := MarshalDatabase(bts)
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

	q := db.Queries.inserts["createUser"]
	if q == nil {
		t.Fatal("createUser is nil")
	}

	st, err := q.Prepare(db)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(st.Statement)
	fmt.Println(st.Args)
	fmt.Println(st.UserInputs)

	fmt.Println("NOW TIME TO PREPARE EXEC")

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

	db, err := MarshalDatabase(bts)
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

	q := db.Queries.updates["updateUser"]
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

	db, err := MarshalDatabase(bts)
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

	q := db.Queries.deletes["removeUser"]
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
