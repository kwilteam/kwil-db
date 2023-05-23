package datasets

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/tx"

	"go.uber.org/zap"
)

func (u *DatasetUseCase) Deploy(ctx context.Context, deployment *entity.DeployDatabase) (*tx.Receipt, error) {
	price, err := u.PriceDeploy(deployment)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(deployment.Tx.Sender, deployment.Tx.Fee, deployment.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	dataset, err := u.engine.NewDataset(ctx, &dto.DatasetContext{
		Name:  deployment.Schema.Name,
		Owner: deployment.Tx.Sender,
	})
	if err != nil {
		return nil, err
	}

	err = deploySchema(ctx, dataset, deployment.Schema)
	if err != nil {
		return nil, err
	}

	u.log.Info("database deployed", zap.String("dbid", dataset.Id()), zap.String("deployer address", deployment.Tx.Sender))

	return &tx.Receipt{
		TxHash: deployment.Tx.Hash,
		Fee:    price.String(),
	}, nil
}

// deploySchema applies the schema to the database
func deploySchema(ctx context.Context, dataset engine.Dataset, schema *entity.Schema) error {
	sp, err := dataset.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	convertedTables, err := convertTablesToDto(schema.Tables)
	if err != nil {
		return err
	}

	for _, table := range convertedTables {
		err := dataset.CreateTable(ctx, table)
		if err != nil {
			return err
		}
	}

	convertedActions, err := convertActionsToDto(schema.Actions)
	if err != nil {
		return err
	}

	for _, action := range convertedActions {
		err := dataset.CreateAction(ctx, action)
		if err != nil {
			return err
		}
	}

	return sp.Commit()
}

func (u *DatasetUseCase) PriceDeploy(deployment *entity.DeployDatabase) (*big.Int, error) {
	return deployPrice, nil
}

func (u *DatasetUseCase) Drop(ctx context.Context, drop *entity.DropDatabase) (*tx.Receipt, error) {
	price, err := u.PriceDrop(drop)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(drop.Tx.Sender, drop.Tx.Fee, drop.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	err = u.engine.DeleteDataset(ctx, &dto.TxContext{
		Caller:  drop.Tx.Sender,
		Dataset: drop.DBID,
	}, drop.DBID)
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
	return dropPrice, nil
}
