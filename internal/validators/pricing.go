package validators

import (
	"context"
	"math/big"
)

var (
	defaultJoinPrice    = big.NewInt(10000000000000)
	defaultLeavePrice   = big.NewInt(0)
	defaultApprovePrice = big.NewInt(0)
	defaultRemovePrice  = big.NewInt(0)
)

// applyFeeMultiplier applies the fee multiplier to the price.
func (mgr *ValidatorMgr) applyFeeMultiplier(price *big.Int) *big.Int {
	return big.NewInt(0).Mul(price, big.NewInt(mgr.feeMultiplier))
}

// PriceJoin returns the price of issuing a join request.
func (mgr *ValidatorMgr) PriceJoin(ctx context.Context) (price *big.Int, err error) {
	return mgr.applyFeeMultiplier(defaultJoinPrice), nil
}

// PriceLeave returns the price of issuing a leave request.
func (mgr *ValidatorMgr) PriceLeave(ctx context.Context) (price *big.Int, err error) {
	return mgr.applyFeeMultiplier(defaultLeavePrice), nil
}

// PriceApprove returns the price of approving a join request.
func (mgr *ValidatorMgr) PriceApprove(ctx context.Context) (price *big.Int, err error) {
	return mgr.applyFeeMultiplier(defaultApprovePrice), nil
}

// PriceRemove returns the price of a remove request.
func (mgr *ValidatorMgr) PriceRemove(ctx context.Context) (price *big.Int, err error) {
	return mgr.applyFeeMultiplier(defaultRemovePrice), nil
}
