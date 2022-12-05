package schema

import (
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

const _YAML_TEST_FILE = "defs_test_simple.yaml"

type TestDefs struct {
	Queries DefinedQueries `yaml:"queries"`
}

//
//func Test_Defs_From_Bytes(t *testing.T) {
//	bts, err := os.ReadFile(_YAML_TEST_FILE)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	defs, err := LoadFromBytes(bts)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	validate_defs(t, defs)
//}
//
//func Test_Defs_From_File(t *testing.T) {
//	defs, err := LoadFromFile(_YAML_TEST_FILE)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	validate_defs(t, defs)
//}
//
//func Test_Defs_From_Map(t *testing.T) {
//	bts, err := os.ReadFile(_YAML_TEST_FILE)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	m := make(map[any]any)
//	err = yaml.Unmarshal(bts, &m)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	defs, err := LoadFromMap(m)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	validate_defs(t, defs)
//}

func Test_Defs_From_Struct(t *testing.T) {
	bts, err := os.ReadFile(_YAML_TEST_FILE)
	if err != nil {
		t.Fatal(err)
	}

	m := Database{}
	err = yaml.Unmarshal(bts, &m)
	if err != nil {
		t.Fatal(err)
	}
}

//
//func validate_defs(t *testing.T, defs map[string]*TableQueryDefs) {
//	if defs["tblUser"] == nil {
//		t.Fatal("tblUser should not be nil")
//	}
//
//	def := defs["tblUser"]
//	if def.Table != "tblUser" {
//		t.Fatal("def.Table should be tblUser")
//	}
//
//	if def.Inserts["createUser"] == nil {
//		t.Fatal("createUser should not be nil")
//	}
//
//	if def.Inserts["createUser"].Name != "createUser" {
//		t.Fatal("def.Inserts[\"createUser\"].Name should be createUser")
//	}
//
//	if def.Updates["updateUser"] == nil {
//		t.Fatal("updateUser should not be nil")
//	}
//
//	if def.Updates["updateUser"].Name != "updateUser" {
//		t.Fatal("def.Updates[\"updateUser\"].Name should be updateUser")
//	}
//
//	if def.Deletes["removeUser"] == nil {
//		t.Fatal("removeUser should not be nil")
//	}
//
//	if def.Deletes["removeUser"].Name != "removeUser" {
//		t.Fatal("def.Deletes[\"removeUser\"].Name should be removeUser")
//	}
//
//	//TODO: add more tests for defs (columns, where, etc)
//}
