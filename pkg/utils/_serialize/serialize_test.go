package serialize_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
)

func Test_Serialize(t *testing.T) {
	marshaler := serialize.Marshaler[MyInter]{
		KnownImplementations: []MyInter{
			MyStruct{},
			MyStruct2{},
		},
	}

	marshaler2 := serialize.Marshaler[MyInter]{
		KnownImplementations: []MyInter{
			MyStruct{},
			MyStruct2{},
		},
	}

	bytes, err := marshaler.Marshal(MyStruct2{Val: "testy", Inter: MyStruct{Val: "testy2"}})
	if err != nil {
		t.Fatal(err)
	}

	retVal, err := marshaler2.Unmarshal(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if retVal.(MyStruct2).Val != "testy" {
		t.Fatal("wrong value")
	}

	if retVal.(MyStruct2).Inter.(MyStruct).Val != "testy2" {
		t.Fatal("wrong value")
	}
}

type MyInter interface {
	Do()
}

type MyStruct struct {
	Val string `json:"val"`
}

func (m MyStruct) Do() {
	println("do")
}

type MyStruct2 struct {
	Val   string  `json:"val"`
	Inter MyInter `json:"inter"`
}

func (m MyStruct2) Do() {
	println("do")
}
