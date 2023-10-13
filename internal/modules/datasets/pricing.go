package datasets

import (
	"context"
	"math/big"

	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
)

var (
	defaultDeployPrice  = big.NewInt(1000000000000000000)
	defaultDropPrice    = big.NewInt(10000000000000)
	defaultExecutePrice = big.NewInt(2000000000000000)
)

// PriceDeploy returns the price of deploying a schema.
func (d *DatasetModule) PriceDeploy(ctx context.Context, schema *engineTypes.Schema) (price *big.Int, err error) {
	return d.applyFeeMultiplier(defaultDeployPrice), nil
}

// PriceDrop returns the price of dropping a schema.
func (d *DatasetModule) PriceDrop(ctx context.Context, dbid string) (price *big.Int, err error) {
	return d.applyFeeMultiplier(defaultDropPrice), nil
}

// PriceExecute returns the price of executing an action.
func (d *DatasetModule) PriceExecute(ctx context.Context, dbid string, action string, args [][]any) (price *big.Int, err error) {
	return d.applyFeeMultiplier(defaultExecutePrice), nil
}

// applyFeeMultiplier applies the fee multiplier to the price.
func (s *DatasetModule) applyFeeMultiplier(price *big.Int) *big.Int {
	return big.NewInt(0).Mul(price, big.NewInt(s.feeMultiplier))
}
