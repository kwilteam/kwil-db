package actions

import (
	"github.com/kwilteam/kwil-db/extensions/actions/mathutil"
	"github.com/kwilteam/kwil-db/extensions/actions/truflation/basestream"
	"github.com/kwilteam/kwil-db/extensions/actions/truflation/stream"
)

func init() {
	err := RegisterExtension("basestream", basestream.InitializeBasestream)
	if err != nil {
		panic(err)
	}

	err = RegisterExtension("truflation_streams", stream.InitializeStream)
	if err != nil {
		panic(err)
	}

	err = RegisterExtension("mathutil", mathutil.InitializeMathUtil)
	if err != nil {
		panic(err)
	}
}
