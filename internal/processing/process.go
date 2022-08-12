package processing

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func (ep *EventProcessor) ProcessEvents(ctx context.Context, ch chan map[string]interface{}) {
	go func() {
		for {
			select {
			case ev := <-ch:
				// Here we can go through and define what we want to do with the event
				switch ev["ktype"] {
				case "Deposit":
					/*
						ev has the following fields:
						    caller (string): the address of the caller
							target (string): the address of the node that the event was sent to
							amount (*big.Int): the amount of the deposit

					*/

					// First, check to ensure that the target is this node's address
					recAddr := ev["target"].(common.Address).String() // convert the target to a string

					// Compare
					if recAddr != ep.Conf.Wallets.Ethereum.Address {
						continue // skip to the next event if the target is not this node's address
					}

					// If we get here, then the target is this node's address
					depAddr := ev["caller"].(common.Address).String() // convert the caller to a string
					err := ep.Deposits.Deposit(ev["amount"].(*big.Int), depAddr)
					if err != nil {
						ep.log.Warn().Err(err).Msg("failed to deposit")
					} else {
						ep.log.Info().Msgf("deposited %s to %s", ev["amount"], depAddr)
					}
				}
			}
		}
	}()
}
