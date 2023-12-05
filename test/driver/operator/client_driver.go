package operator

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/adminclient"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/test/driver"
)

type AdminClientDriver struct {
	Client *adminclient.AdminClient
}

var _ KwilOperatorDriver = (*AdminClientDriver)(nil)

func (a *AdminClientDriver) TxSuccess(ctx context.Context, txHash []byte) error {
	resp, err := a.Client.TxQuery(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	if resp.TxResult.Code != transactions.CodeOk.Uint32() {
		return fmt.Errorf("transaction not ok, %s", resp.TxResult.Log)
	}

	// NOTE: THIS should not be considered a failure, should retry
	if resp.Height < 0 {
		return driver.ErrTxNotConfirmed
	}

	return nil
}

func (a *AdminClientDriver) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error) {
	return a.Client.JoinStatus(ctx, pubKey)
}

func (a *AdminClientDriver) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error) {
	return a.Client.Approve(ctx, joinerPubKey)
}

func (a *AdminClientDriver) ValidatorNodeJoin(ctx context.Context) ([]byte, error) {
	return a.Client.Join(ctx)
}

func (a *AdminClientDriver) ValidatorNodeLeave(ctx context.Context) ([]byte, error) {
	return a.Client.Leave(ctx)
}

func (a *AdminClientDriver) ValidatorNodeRemove(ctx context.Context, target []byte) ([]byte, error) {
	return a.Client.Remove(ctx, target)
}

func (a *AdminClientDriver) ValidatorsList(ctx context.Context) ([]*types.Validator, error) {
	return a.Client.ListValidators(ctx)
}
