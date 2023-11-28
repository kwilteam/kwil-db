package escrow

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

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
		fmt.Println("Deposit event found: ", edi.Event.Caller.Hex(), edi.Event.Receiver.Hex(), edi.Event.Amount.String())
		receiver := strings.ToLower(edi.Event.Receiver.Hex())
		if receiver != strings.ToLower(escrow.escrowAddr) {
			fmt.Println("receiver is not escrow address", receiver, " expected: ", escrow.escrowAddr) // TODO: Use logger
			continue
		}

		// Unique ID for the deposit: hash(sender + amount + txHash + blockHash + ChainID)
		hasher := sha256.New()
		hasher.Write([]byte("Deposit"))
		hasher.Write([]byte(edi.Event.Caller.Hex()))
		hasher.Write([]byte(edi.Event.Amount.String()))
		hasher.Write([]byte(edi.Event.Raw.TxHash.Hex()))
		hasher.Write([]byte(edi.Event.Raw.BlockHash.Hex()))
		hasher.Write([]byte(escrow.chainId.String()))
		id := hasher.Sum(nil)

		deposit := &chain.DepositEvent{
			ID:     fmt.Sprintf("%x", id),
			Sender: edi.Event.Caller.Hex(),
			Amount: edi.Event.Amount,
		}
		deposits = append(deposits, deposit)
		fmt.Printf("Deposit: %+v\n", deposit)
	}
	return deposits
}
