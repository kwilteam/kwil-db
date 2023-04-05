package datasets

import (
	"kwil/internal/entity"
	"kwil/pkg/tx"
	"math/big"

	"go.uber.org/zap"
)

func (u *DatasetUseCase) Deploy(deployment *entity.DeployDatabase) (*tx.Receipt, error) {
	price, err := u.PriceDeploy(deployment)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(deployment.Tx.Sender, deployment.Tx.Fee, deployment.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	err = u.engine.Deploy(deployment.Schema)
	if err != nil {
		return nil, err
	}

	u.log.Info("database deployed", zap.String("dbid", deployment.Schema.ID()), zap.String("deployer address", deployment.Tx.Sender))

	return &tx.Receipt{
		TxHash: deployment.Tx.Hash,
		Fee:    price.String(),
	}, nil
}

func (u *DatasetUseCase) PriceDeploy(deployment *entity.DeployDatabase) (*big.Int, error) {
	return u.engine.GetDeployPrice(deployment.Schema)
}

func (u *DatasetUseCase) Drop(drop *entity.DropDatabase) (*tx.Receipt, error) {
	price, err := u.PriceDrop(drop)
	if err != nil {
		return nil, err
	}

	err = u.compareAndSpend(drop.Tx.Sender, drop.Tx.Fee, drop.Tx.Nonce, price)
	if err != nil {
		return nil, err
	}

	err = u.engine.DropDataset(drop.DBID)
	if err != nil {
		return nil, err
	}

	return &tx.Receipt{
		TxHash: drop.Tx.Hash,
		Fee:    price.String(),
	}, nil
}

func (u *DatasetUseCase) PriceDrop(drop *entity.DropDatabase) (*big.Int, error) {
	return u.engine.GetDropPrice(drop.DBID)
}
