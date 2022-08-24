package events

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"math/big"
)

// This function takes a channel of block heights and returns a channel of events.
func (ef *EventFeed) processBlocks(ctx context.Context, ch chan *big.Int) {
	// This method is essentially public since it is one of two methods used in ef.Listen
	addr := common.HexToAddress(ef.Config.ClientChain.DepositContract.Address)
	go func() {
		for {
			// At this point, we have received a finalized ethereum block
			// Now to query the event data
			height := <-ch
			ef.logHeight(height)

			// TODO: We should probably have some retry logic here for transient unavailability.
			query := ethereum.FilterQuery{
				FromBlock: height,
				ToBlock:   height,
				Addresses: []common.Address{addr},
				Topics:    [][]common.Hash{ef.getTopicsForEvents()},
			}

			// Get a channel that will return the events
			logs, err := ef.EthClient.FilterLogs(ctx, query)
			if err != nil {
				ef.log.Fatal().Err(err).Msg("error reading in block data")
			}

			for i := 0; i < len(logs); i++ {
				err = ef.ProcessLog(logs[i])
				if err != nil {
					ef.log.Error().Err(err).Msg("error processing log")
				}
			}

			// At this point, we have confirmed stored all changes for the block, and can now delete any of the txs stored in the deposit store
			ef.ds.CommitBlock(height)
		}
	}()
}

func (ed *EventFeed) logHeight(h *big.Int) {
	bi := big.NewInt(0)
	if bi.Mod(h, big.NewInt(1)).Cmp(big.NewInt(0)) == 0 {
		log.Debug().Msgf("Processing block %d", h)
	}
}
