package escrow

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	escrowAbi "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/escrow/abi"
	"github.com/kwilteam/kwil-db/core/types/chain"
)

func (e *Escrow) GetDeposits(ctx context.Context, from uint64, to *uint64) ([]*chain.DepositEvent, error) {
	queryOpts := &bind.FilterOpts{Context: ctx, Start: from, End: to}

	iter, err := e.ctr.FilterDeposit(queryOpts)
	if err != nil {
		return nil, err
	}

	return e.retrieveDeposits(iter, e.tokenAddr), nil
}

func (escrow *Escrow) retrieveDeposits(edi *escrowAbi.EscrowDepositIterator, token string) []*chain.DepositEvent {
	var deposits []*chain.DepositEvent

	for edi.Next() {
		fmt.Println("Deposit event found: ", edi.Event)
		// receiver := edi.Event.Receiver.Hex()
		// if receiver != escrow.escrowAddr {
		// 	fmt.Println("receiver is not escrow address") // TODO: Use logger
		// 	continue
		// }
		deposit := &chain.DepositEvent{
			Sender:    edi.Event.Caller.Hex(),
			Receiver:  edi.Event.Receiver.Hex(),
			Amount:    edi.Event.Amount.String(),
			Height:    int64(edi.Event.Raw.BlockNumber),
			TxHash:    edi.Event.Raw.TxHash.Hex(),
			BlockHash: edi.Event.Raw.BlockHash.Hex(),
		}
		deposits = append(deposits, deposit)
	}
	return deposits
}
