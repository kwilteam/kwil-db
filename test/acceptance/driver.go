package acceptance

import (
	"kwil/test/specifications"
)

type KwilAcceptanceDriver interface {
	specifications.DatabaseDeployDsl
	specifications.DatabaseDropDsl
	specifications.ApproveTokenDsl
	specifications.DepositFundDsl
	specifications.ExecuteQueryDsl
}
