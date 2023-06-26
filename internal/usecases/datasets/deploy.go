package datasets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/tx"
	"github.com/pkg/errors"

	"go.uber.org/zap"
)

func (u *DatasetUseCase) Deploy(ctx context.Context, deployment *entity.DeployDatabase) (rec *tx.Receipt, err error) {
	price := big.NewInt(0)

	if u.gas_enabled {
		price, err = u.PriceDeploy(deployment)
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Tx fee: %v  Gas Price: %s\n", deployment.Tx.Fee, price)
	err = u.compareAndSpend(deployment.Tx.Sender, deployment.Tx.Fee, deployment.Tx.Nonce, price)
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
	dataset, err := u.engine.NewDataset(ctx, &dto.DatasetContext{
		Name:  deployment.Schema.Name,
		Owner: deployment.Tx.Sender,
	})
	if err != nil {
		return err
	}

	err = u.deploySchema(ctx, dataset, deployment.Schema)
	if err != nil {
		err2 := u.engine.DeleteDataset(ctx, &dto.TxContext{
			Caller: deployment.Tx.Sender,
		}, dataset.Id())

		if err2 != nil {
			u.log.Error("failed to delete dataset after failed schema deployment", zap.Error(err2))
			err = errors.Wrap(err, err2.Error())
		}

		return err
	}

	u.log.Info("database deployed", zap.String("dbid", dataset.Id()), zap.String("deployer address", deployment.Tx.Sender))

	return nil
}

// deploySchema applies the schema to the database
func (u *DatasetUseCase) deploySchema(ctx context.Context, dataset engine.Dataset, schema *entity.Schema) error {
	convertedTables, err := convertTablesToDto(schema.Tables)
	if err != nil {
		return err
	}

	sp, err := dataset.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

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

func (u *DatasetUseCase) Drop(ctx context.Context, drop *entity.DropDatabase) (txReceipt *tx.Receipt, err error) {
	// TODO: there are a lot of errors with drop and having potentially orphaned data
	// this can cause panics.  For now, I will catch panics since we are releasing today
	defer func() {
		if r := recover(); r != nil {
			u.log.Error("recovering from panic in drop", zap.Any("panic", r))
			err = errors.New("Unexpected internal error. Please report this to the Kwil team.")
		}
	}()
	price := big.NewInt(0)

	if u.gas_enabled {
		price, err = u.PriceDrop(drop)
		if err != nil {
			return nil, err
		}
	}
	fmt.Printf("Tx fee: %v  Gas Price: %s\n", drop.Tx.Fee, price)
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
