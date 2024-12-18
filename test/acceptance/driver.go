package acceptance

import (
	"github.com/kwilteam/kwil-db/test/specifications"
)

type KwilAcceptanceDriver interface {
	specifications.ExecuteQueryDsl
	specifications.ExecuteCallDsl
	specifications.InfoDsl
	specifications.AccountBalanceDsl
	specifications.TransferAmountDsl
	specifications.TxInfoer
}
