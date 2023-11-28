package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/test/integration"
	"github.com/kwilteam/kwil-db/test/specifications"
)

/*
 Integration tests for the token bridge scenarios:
	WithoutGasCosts: false, TokenBridge enabled

	Things to test:
	-
*/

func TestTokenBridge(t *testing.T) {
	// TokenBridge Approvals & Deposits

	// TokenBridge Transactions (Deploy, Drop, Execute, Call, maybe Validator Changes)
	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(4),
		integration.WithNonValidators(0),
		integration.WithNumConfirmations(0),
		integration.WithoutGasCosts(false),
	}

	helper := integration.NewIntHelper(t, opts...)
	helper.Setup(ctx, allServices)
	defer helper.Teardown()

	// running forever for local development
	if *dev {
		helper.WaitForSignals(t)
		return
	}

	userDriver := helper.GetUserDriver(ctx, "node0", "client")

	spender := helper.EscrowAddress()

	specifications.TokenBridgeApproveSuccessSpecification(ctx, t, spender, userDriver)

	specifications.TokenBridgeDepositSuccessSpecification(ctx, t, spender, userDriver)
}
