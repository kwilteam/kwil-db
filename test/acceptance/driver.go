package acceptance

import (
	"kwil/test/specifications"
)

type KwilACTDriver interface {
	specifications.DatabaseDeployDsl
	specifications.DatabaseDropDsl
	specifications.ApproveTokenDsl
	specifications.DepositFundDsl
	specifications.ExecuteQueryDsl
}
