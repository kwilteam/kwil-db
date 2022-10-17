package contracts

import (
	"context"

	"kwil/abi"
	ct "kwil/x/deposits/chainclient/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type deposit struct {
	ed    *abi.EscrowDeposit
	token *string
}

func (c *contract) GetDeposits(ctx context.Context, from, to int64) ([]ct.Deposit, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	edi, err := c.ctr.FilterDeposit(queryOpts)
	if err != nil {
		return nil, err
	}

	return convertDeposits(edi, &c.token), nil
}

func (d *deposit) Caller() string {
	return d.ed.Caller.Hex()
}

func (d *deposit) Target() string {
	return d.ed.Target.Hex()
}

func (d *deposit) Amount() string {
	return d.ed.Amount.String()
}

func (d *deposit) Height() int64 {
	return int64(d.ed.Raw.BlockNumber)
}

func (d *deposit) Tx() string {
	return d.ed.Raw.TxHash.Hex()
}

func (d *deposit) Type() uint8 {
	return 0
}

func (d *deposit) Token() string {
	return *d.token
}

func convertDeposits(edi *abi.EscrowDepositIterator, token *string) []ct.Deposit {
	var deposits []ct.Deposit
	for {
		deposits = append(deposits, escToDeposit(edi.Event, token))
		if !edi.Next() {
			break
		}
	}

	return deposits
}

// escToDeposit converts abi.EscrowDeposit to ct.Deposit
func escToDeposit(ed *abi.EscrowDeposit, token *string) ct.Deposit {
	return &deposit{
		ed:    ed,
		token: token,
	}
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
