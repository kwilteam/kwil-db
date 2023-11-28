package deployer

import (
	"context"
)

type Deployer interface {
	DeployEscrow(ctx context.Context, tokenAddr string) (string, error)
	DeployToken(ctx context.Context) (string, error)
	// FundAccount(ctx context.Context, account string, amount int64) error
}
