package events

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func (ef *EventFeed) GetUnsyncedRange(ctx context.Context) (*big.Int, *big.Int, error) {
	lowH, err := ef.ds.GetLastHeight()
	if err != nil {
		return nil, nil, err
	}

	// Get the current height from Ethereum
	curHead, err := ef.EthClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

<<<<<<< HEAD
	highH := curHead.Number.Sub(curHead.Number, big.NewInt(int64(ef.conf.GetReqConfirmations()))) // Subtracting the max buffer size from the current height
=======
	highH := curHead.Number.Sub(curHead.Number, big.NewInt(int64(ef.Config.ClientChain.RequiredConfirmations))) // Subtracting the max buffer size from the current height
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	// Now we run through all heights from lowH to highH
	return lowH, highH, nil
}

func (ef *EventFeed) GetUnsyncedEvents(ctx context.Context, low *big.Int, high *big.Int) ([]Event, error) {

	// Define event array
	evs := []Event{}
	// Get contract address
<<<<<<< HEAD
	addr := common.HexToAddress(ef.conf.GetDepositAddress())
=======
	addr := common.HexToAddress(ef.Config.ClientChain.DepositContract.Address)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

	// Make query based on params
	filter := ethereum.FilterQuery{
		FromBlock: low,
		ToBlock:   high, // Add 1 to the high because we want to include the last block
		Addresses: []common.Address{addr},
		Topics:    [][]common.Hash{ef.getTopicsForEvents()},
	}

	// Filter the logs
	logs, err := ef.EthClient.FilterLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Loops through the logs
	for _, vLog := range logs {
		ev, err := ef.parseEvent(vLog)
		if err != nil {
			ef.log.Fatal().Err(err).Msg("error parsing event")
		}
		evs = append(evs, ev)
	}

	return evs, nil
}

func (ef *EventFeed) IndicateLastHeight() error {
	lastHeight, err := ef.ds.GetLastHeight()
	if err != nil {
		return err
	}

	// This is 0 in case it somehow does not get updated
<<<<<<< HEAD
	if lastHeight.Cmp(big.NewInt(int64(ef.conf.GetLowestHeight()))) >= 0 { // Comparing the lowest tracked height to the stored height
=======
	if lastHeight.Cmp(big.NewInt(int64(ef.Config.ClientChain.LowestHeight))) >= 0 { // Comparing the lowest tracked height to the stored height
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
		// This means that the deposit store height is greater than the lowest possible.
		// We should use the stored height to sync from.
		ef.log.Debug().Msgf("using stored height to sync from: %s", lastHeight.String())
		lowH := big.NewInt(lastHeight.Int64())
		err = ef.ds.SetLastHeight(lowH)
		if err != nil {
			return err
		}
	} else {
		// This means that the deposit store height is less than the lowest possible.
		// We should use the lowest possible height to sync from.
<<<<<<< HEAD
		ef.log.Debug().Msgf("using lowest possible height to sync from: %d", ef.conf.GetLowestHeight())
		lowH := big.NewInt(int64(ef.conf.GetLowestHeight()))
=======
		ef.log.Debug().Msgf("using lowest possible height to sync from: %d", ef.Config.ClientChain.LowestHeight)
		lowH := big.NewInt(int64(ef.Config.ClientChain.LowestHeight))
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
		err = ef.ds.SetLastHeight(lowH)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ef *EventFeed) UpdateLastHeight(h *big.Int) error {
	return ef.ds.SetLastHeight(h)
}
