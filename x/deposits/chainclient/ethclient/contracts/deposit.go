package contracts

import (
	"context"

	"kwil/abi"
	ct "kwil/x/deposits/chainclient/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func (c *contract) GetDeposits(ctx context.Context, from, to int64) ([]*ct.Deposit, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	edi, err := c.ctr.FilterDeposit(queryOpts)
	if err != nil {
		return nil, err
	}

	return convertDeposits(edi, c.token), nil
}

func convertDeposits(edi *abi.EscrowDepositIterator, token string) []*ct.Deposit {
	var deposits []*ct.Deposit
	for {

		if !edi.Next() {
			break
		} else {
			deposits = append(deposits, escToDeposit(edi.Event, token))
		}
	}

	return deposits
}

// escToDeposit converts abi.EscrowDeposit to deposit
func escToDeposit(ed *abi.EscrowDeposit, token string) *ct.Deposit {
	// print all fields

	return ct.NewDeposit(ed.Caller.Hex(), ed.Target.Hex(), ed.Amount.String(), int64(ed.Raw.BlockNumber), ed.Raw.TxHash.Hex(), 0, token)
}

// we don't need this anymore
/*func (c *contract) SubDepositEvents(ctx context.Context) (event.Subscription, chan<- *abi.EscrowDeposit, error) {
	watchOpts := &bind.WatchOpts{Context: ctx, Start: nil}

	ch := make(chan *abi.EscrowDeposit)
	sub, err := c.ctr.WatchDeposit(watchOpts, ch)
	if err != nil {
		return nil, nil, err
	}

	return sub, ch, nil
}*/
