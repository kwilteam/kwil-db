package tsts_test

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/tsts"
)

func Test_Load(t *testing.T) {
	val := tsts.GetFromJson[types.Table]("likes")

	fmt.Println(val)
}
