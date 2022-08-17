package deposits

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/internal/events"
	"github.com/kwilteam/kwil-db/internal/store"
	"github.com/kwilteam/kwil-db/internal/wal"
	"github.com/kwilteam/kwil-db/pkg/types"
	//"math/big"
)

/*
	This package pulls in the event feed, event processor, and deposit store.
	The main reason I separated this from main.go is since it will get a little messy sharing things like the WAL between all 3
*/

type Deposits struct {
	Store *store.DepositStore
	ef    *events.EventFeed
	wal   *types.Wal
}

/*
	Constructor needs:
		- ctx
		- config
		- eth client
*/

const walPath = ".wal"

func Init(ctx context.Context, conf *types.Config, client *ethclient.Client) (*Deposits, error) {

	// Make a WAL
	wal, err := wal.OpenEthTxWal(walPath)
	if err != nil {
		return nil, err
	}

	// First initialize a deposit store
	ds, err := store.NewDepositStore(conf, wal)
	ds.PrintAllBalances()
	ds.PrintCurrentHeight()
	if err != nil {
		return nil, err
	}

	// Next initialize an event feed
	ef, err := events.New(conf, client, wal, ds)
	if err != nil {
		return nil, err
	}

	// Make sure that the height is properly set
	err = ef.IndicateLastHeight()
	if err != nil {
		return nil, err
	}

	// Now, we sync with old events
	err = Sync(ctx, ef)
	if err != nil {
		return nil, err
	}

	// Next start the event feed
	err = ef.Listen(ctx)
	if err != nil {
		return nil, err
	}

	return &Deposits{
		Store: ds,
		ef:    ef,
		wal:   getWalPtr(wal),
	}, nil
}

func Sync(ctx context.Context, ef *events.EventFeed) error {
	fmt.Printf("Beginning sync...\n")
	low, high, err := ef.GetUnsyncedRange(ctx)
	//high = high.Add(high, big.NewInt(1))
	if err != nil {
		return err
	}
	fmt.Printf("Syncing from %d to %d\n", low, high)

	events, err := ef.GetUnsyncedEvents(ctx, low, high)
	if err != nil {
		return err
	}

	// Iterate over all events and process them
	for _, ev := range events {

		err = ef.ProcessEvent(ev)
		if err != nil {
			return err
		}
	}

	// Finally, update the last height
	err = ef.UpdateLastHeight(high)
	if err != nil {
		return err
	}

	fmt.Printf("Sync complete at height %d!\n", high)
	return nil
}

func getWalPtr(wal types.Wal) *types.Wal {
	return &wal
}
