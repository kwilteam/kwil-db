package contracts

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"kwil/abi"
	ct "kwil/x/deposits/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) GetDeposits(ctx context.Context, from, to int64, addr string) ([]*ct.Deposit, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	ads := common.HexToAddress(addr)

	edi, err := c.ctr.FilterDeposit(queryOpts, []common.Address{ads})
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

// RetrunFunds calls the returnDeposit function
func (c *contract) ReturnFunds(ctx context.Context, key *ecdsa.PrivateKey, recipient, nonce string, amt *big.Int, fee *big.Int) error {

	txOpts, err := bind.NewKeyedTransactorWithChainID(key, c.cid)
	if err != nil {
		return err
	}

	_, err = c.ctr.ReturnDeposit(txOpts, common.HexToAddress(recipient), amt, fee, nonce)
	if err != nil {
		return err
	}

	return nil
}
