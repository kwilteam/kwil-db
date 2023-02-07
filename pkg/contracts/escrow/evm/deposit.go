package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/escrow/evm/abi"
	escrow2 "kwil/pkg/types/contracts/escrow"
)

func (c *contract) GetDeposits(ctx context.Context, from, to int64) ([]*escrow2.DepositEvent, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	ads := common.HexToAddress(c.nodeAddress)

	edi, err := c.ctr.FilterDeposit(queryOpts, []common.Address{ads})
	if err != nil {
		return nil, err
	}

	return convertDeposits(edi, c.token), nil
}

func convertDeposits(edi *abi.EscrowDepositIterator, token string) []*escrow2.DepositEvent {
	var deposits []*escrow2.DepositEvent
	for {

		if !edi.Next() {
			break
		} else {
			deposits = append(deposits, &escrow2.DepositEvent{
				Caller: edi.Event.Caller.Hex(),
				Target: edi.Event.Target.Hex(),
				Amount: edi.Event.Amount.String(),
				Height: int64(edi.Event.Raw.BlockNumber),
				TxHash: edi.Event.Raw.TxHash.Hex(),
			})
		}
	}

	return deposits
}

func (c *contract) Deposit(ctx context.Context, params *escrow2.DepositParams) (*escrow2.DepositResponse, error) {

	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, c.privateKey)
	if err != nil {
		return nil, err
	}

	res, err := c.ctr.Deposit(auth, common.HexToAddress(params.Validator), params.Amount)
	if err != nil {
		return nil, err
	}

	return &escrow2.DepositResponse{
		TxHash: res.Hash().String(),
	}, nil
}
