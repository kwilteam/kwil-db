package integration

import (
	"github.com/kwilteam/kwil-db/test/specifications"
)

type KwilIntDriver interface {
	specifications.ExecuteQueryDsl
	specifications.ExecuteCallDsl
	specifications.TransferAmountDsl
	specifications.DeployerDsl
}
