package evm

import (
	"context"
	"crypto/ecdsa"
	kwilCommon "kwil/pkg/contracts/common/evm"
	"kwil/pkg/contracts/escrow/evm/abi"
	"kwil/pkg/contracts/escrow/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) GetDeposits(ctx context.Context, from, to int64, providerAddress string) ([]*types.DepositEvent, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	ads := common.HexToAddress(providerAddress)

	edi, err := c.ctr.FilterDeposit(queryOpts, []common.Address{ads})
	if err != nil {
		return nil, err
	}

	return convertDeposits(edi, c.token), nil
}

func convertDeposits(edi *abi.EscrowDepositIterator, token string) []*types.DepositEvent {
	var deposits []*types.DepositEvent
	for {

		if !edi.Next() {
			break
		} else {
			deposits = append(deposits, &types.DepositEvent{
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

func (c *contract) Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error) {

	auth, err := kwilCommon.PrepareTxAuth(ctx, c.client, c.chainId, privateKey)
	if err != nil {
		return nil, err
	}

	res, err := c.ctr.Deposit(auth, common.HexToAddress(params.Validator), params.Amount)
	if err != nil {
		return nil, err
	}

	return &types.DepositResponse{
		TxHash: res.Hash().String(),
	}, nil
}
