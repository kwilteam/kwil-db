package datasets

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/tx"

	"go.uber.org/zap"
)

func (u *DatasetUseCase) Deploy(ctx context.Context, deployment *entity.DeployDatabase) (rec *tx.Receipt, err error) {
	price, err := u.PriceDeploy(deployment)
	if err != nil {
		return nil, err
	}

	err = u.CompareAndSpend(deployment.Tx.Sender, deployment.Tx.Fee, deployment.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	err = u.deployDataset(ctx, deployment)
	if err != nil {
		return nil, err
	}

	return &tx.Receipt{
		TxHash: deployment.Tx.Hash,
		Fee:    price.String(),
	}, nil
}

func (u *DatasetUseCase) deployDataset(ctx context.Context, deployment *entity.DeployDatabase) error {
	tables, err := convertTablesToDto(deployment.Schema.Tables)
	if err != nil {
		return err
	}

	actions, err := convertActionsToDto(deployment.Schema.Actions)
	if err != nil {
		return err
	}

	extensions := convertExtensionsToDto(deployment.Schema.Extensions)

	dbid, err := u.engine.CreateDataset(ctx, deployment.Schema.Name, deployment.Schema.Owner, &engine.Schema{
		Tables:     tables,
		Procedures: actions,
		Extensions: extensions,
	})
	if err != nil {
		return err
	}

	u.log.Info("database deployed", zap.String("dbid", dbid), zap.String("deployer address", deployment.Tx.Sender))

	return nil
}

func (u *DatasetUseCase) PriceDeploy(deployment *entity.DeployDatabase) (*big.Int, error) {
	if u.accountStore.GasEnabled() {
		return deployPrice, nil
	}
	return big.NewInt(0), nil
}

func (u *DatasetUseCase) Drop(ctx context.Context, drop *entity.DropDatabase) (txReceipt *tx.Receipt, err error) {
	price, err := u.PriceDrop(drop)
	if err != nil {
		return nil, err
	}

	err = u.CompareAndSpend(drop.Tx.Sender, drop.Tx.Fee, drop.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	err = u.engine.DropDataset(ctx, drop.Tx.Sender, drop.DBID)
	if err != nil {
		return nil, err
	}

	u.log.Info("database dropped", zap.String("dbid", drop.DBID), zap.String("dropper address", drop.Tx.Sender))

	return &tx.Receipt{
		TxHash: drop.Tx.Hash,
		Fee:    price.String(),
	}, nil
}

func (u *DatasetUseCase) PriceDrop(drop *entity.DropDatabase) (*big.Int, error) {
	if u.accountStore.GasEnabled() {
		return dropPrice, nil
	}
	return big.NewInt(0), nil
}
