package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"kwil/pkg/contracts/escrow/evm/abi"
	"kwil/pkg/contracts/escrow/types"
)

func (c *contract) GetWithdrawals(ctx context.Context, from, to int64) ([]*types.WithdrawalConfirmationEvent, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	address := common.HexToAddress(c.nodeAddress)

	edi, err := c.ctr.FilterWithdrawal(queryOpts, []common.Address{address})
	if err != nil {
		return nil, err
	}

	return convertWithdrawals(edi, c.token), nil
}

func convertWithdrawals(edi *abi.EscrowWithdrawalIterator, token string) []*types.WithdrawalConfirmationEvent {
	var withdrawals []*types.WithdrawalConfirmationEvent
	for {

		if !edi.Next() {
			break
		} else {
			withdrawals = append(withdrawals, &types.WithdrawalConfirmationEvent{
				Caller:   edi.Event.Caller.Hex(),   // this is the node address / this machine
				Receiver: edi.Event.Receiver.Hex(), // this is the wallet that we returned the funds to
				Amount:   edi.Event.Amount.String(),
				Fee:      edi.Event.Fee.String(),
				Cid:      edi.Event.Nonce,
				Height:   int64(edi.Event.Raw.BlockNumber),
				TxHash:   edi.Event.Raw.TxHash.Hex(),
			})
		}
	}

	return withdrawals
}
