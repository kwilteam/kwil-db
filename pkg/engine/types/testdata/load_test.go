package testdata_test

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
)

func Test_Load(t *testing.T) {
	val := testdata.GetFromJson[types.Table]("likes")

	fmt.Println(val)
}
