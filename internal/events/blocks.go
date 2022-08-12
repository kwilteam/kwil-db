package events

import (
	"context"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/internal/config"
	kwal "github.com/kwilteam/kwil-db/internal/wal"
	"github.com/rs/zerolog/log"
)

type WAL interface{}

const walPref = "events"

// This function takes a channel of block heights and returns a channel of events.
func (e *EventFeed) pullEvents(ctx context.Context, ch chan *big.Int) chan map[string]interface{} {
	retChan := make(chan map[string]interface{})
	go func() {
		for {
			// At this point, we have received a finalized ethereum block
			// Now to query the event data
			height := <-ch

			// Create new WAL here
			wal, err := kwal.Open(path.Join(walPref, height.String()))
			if err != nil {
				log.Fatal().Err(err).Msg("error creating new wal")
			}

			// Append BeginBlock
			err = wal.BeginEthBlock(height)
			if err != nil {
				log.Fatal().Err(err).Msg("error writing BeginEthBlock to wal")
			}

			addr := common.HexToAddress(config.Conf.ClientChain.DepositContract.Address)

			// TODO: We should probably have some retry logic here for transient unavailability.  I haven't seen it yet and have let it run for a while.
			query := ethereum.FilterQuery{
				FromBlock: height,
				ToBlock:   height,
				Addresses: []common.Address{addr},
				Topics:    [][]common.Hash{e.getTopicsForEvents()},
			}

			// Get a channel that will return the events
			logs, err := e.EthClient.FilterLogs(ctx, query)
			if err != nil {
				log.Fatal().Err(err).Msg("error reading in block data")
			}

			for _, vLog := range logs {

				// Append tx hash to wal
				hBytes := common.Hex2BytesFixed(vLog.TxHash.Hex(), 32)
				err = wal.BeginTransaction(hBytes)
				if err != nil {
					log.Fatal().Err(err).Msg("error writing BeginTransaction to wal")
				}

				// First I will find the topic
				topic := vLog.Topics[0]
				// Next, I find the event name
				event := e.Topics[topic]

				// Next, I will unpack based on the event
				ev, err := e.ClientChain.GetContractABI().Unpack(event.Name, vLog.Data)
				if err != nil {
					log.Fatal().Err(err).Msg("error unpacking event data")
				}
				if len(ev) != len(event.Inputs) {
					log.Fatal().Err(err).Msg("received smart contract event with different number of inputs than expected")
				}

				// Create a map to store the event data with dynamic keys
				em := map[string]interface{}{}
				em["ktype"] = event.Name // I name this ktype to ensure there aren't collisions with other fields
				// Loop over the inputs and add them to the map
				// This allows us to dyanmically name fields based on the ABI
				for i := 0; i < len(ev); i++ {
					// Get the name for the first arg
					name := event.Inputs[i].Name
					em[name] = ev[i]
				}
				// Send the result through the channel
				retChan <- em
			}
			// Append EndBlock
			err = wal.EndEthBlock(height)
			if err != nil {
				log.Fatal().Err(err).Msg("error writing EndEthBlock to wal")
			}

			// Close the wal
			//err = wal.Close()
			if err != nil {
				log.Warn().Err(err).Msg("error closing wal")
			}
		}
	}()
	return retChan
}
