package escrow

import (
	"context"
	"kwil/abi"
	"kwil/x/deposits/dto"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) GetDeposits(ctx context.Context, from, to int64) ([]*dto.Deposit, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	ads := common.HexToAddress(c.nodeAddress)

	edi, err := c.ctr.FilterDeposit(queryOpts, []common.Address{ads})
	if err != nil {
		return nil, err
	}

	return convertDeposits(edi, c.token), nil
}

func convertDeposits(edi *abi.EscrowDepositIterator, token string) []*dto.Deposit {
	var deposits []*dto.Deposit
	for {

		if !edi.Next() {
			break
		} else {
			deposits = append(deposits, &dto.Deposit{
				Wallet: edi.Event.Caller.Hex(),
				Amount: edi.Event.Amount.String(),
				Height: int64(edi.Event.Raw.BlockNumber),
				TxHash: edi.Event.Raw.TxHash.Hex(),
			})
		}
	}

	return deposits
}
