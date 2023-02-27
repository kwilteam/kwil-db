package deployer

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
)

type Deployer interface {
	DeployEscrow(ctx context.Context, tokenAddr string) (common.Address, error)
	DeployToken(ctx context.Context) (common.Address, error)
	FundAccount(ctx context.Context, account string, amount int64) error
}
