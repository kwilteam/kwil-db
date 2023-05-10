package traverse_test

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models/clean/traverse"
	"reflect"
	"testing"
)

func Test_Traverse(t *testing.T) {
	a := MyStruct{
		Name: "name",
		Age:  10,
		InnerStruct: InnerStruct{
			InnnerName: "innerName",
		},
		Arrs: []string{"arr1", "arr2"},
	}

	traverser := traverse.New("tag1", func(field reflect.Value, tags []string) {
		fmt.Println(tags)
		if tags[0] == "value1" {
			field.SetString("NEWNAME")
		}
	})

	traverser.Traverse(&a)

	if a.Name != "NEWNAME" {
		t.Errorf("Name should be NEWNAME")
	}
}

type MyStruct struct {
	Name        string `tag1:"value1" tag2:"value2"`
	Age         int    `tag1:"value3" tag2:"value4"`
	InnerStruct InnerStruct
	Arrs        []string `tag1:"value5" tag2:"value6"`
}

type InnerStruct struct {
	InnnerName string `tag1:"value7" tag2:"value8"`
}
