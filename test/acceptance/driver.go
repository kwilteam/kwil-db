package acceptance

import (
	"github.com/kwilteam/kwil-db/test/specifications"
)

type KwilAcceptanceDriver interface {
	specifications.DatabaseDeployDsl
	specifications.DatabaseDropDsl
	specifications.ExecuteQueryDsl
	specifications.ExecuteCallDsl
	specifications.ValidatorOpsDsl
}
