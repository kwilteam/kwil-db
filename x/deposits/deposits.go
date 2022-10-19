package deposits

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"kwil/x/cfgx"
	ct "kwil/x/deposits/chainclient/types"
	"kwil/x/deposits/events"
	"kwil/x/deposits/store"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Deposits interface {
	Listen(context.Context) error
	GetBalance(string) (*big.Int, error)
	GetSpent(string) (*big.Int, error)
	Spend(string, *big.Int) error
	Close() error
}

type deposits struct {
	log  *zerolog.Logger
	conf cfgx.Config
	ef   events.EventFeed
	sc   ct.Contract
	lh   int64
	ds   store.DepositStore
	// TODO: add keyring
}

func New(c cfgx.Config) (*deposits, error) {
	logger := log.With().Str("module", "deposits").Logger()

	ds, err := store.New(c)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize deposit store. %w", err)
	}

	lb, err := ds.GetLastHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block height. %w", err)
	}

	ef, err := events.New(c, lb)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event feed. %w", err)
	}

	return &deposits{
		log:  &logger,
		conf: c,
		ef:   ef,
		sc:   ef.Contract(),
		lh:   lb,
		ds:   ds,
	}, nil
}

func (d *deposits) Listen(ctx context.Context) error {

	blks, errs, err := d.ef.Listen(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			return err
		case blk := <-blks:
			d.processBlock(ctx, blk)
		}
	}
}

func (d *deposits) processBlock(ctx context.Context, blk int64) {
	d.log.Debug().Int64("height", blk).Msg("processing block")

	wg := sync.WaitGroup{}
	wg.Add(2) // processing deposits and withdrawals simultaneously

	go func(context.Context, int64) {
		// process deposits
		err := d.processDeposits(ctx, blk)
		if err != nil {
			d.log.Warn().Err(err).Int64("height", blk).Msg("failed to process deposits")
		}
		wg.Done()
	}(ctx, blk)

	go func(context.Context, int64) {
		// process withdrawals
		err := d.processWithdrawals(ctx, blk)
		if err != nil {
			d.log.Warn().Err(err).Int64("height", blk).Msg("failed to process withdrawals")
		}
		wg.Done()
	}(ctx, blk)
}

func (d *deposits) processDeposits(ctx context.Context, blk int64) error {
	d.log.Debug().Int64("height", blk).Msg("processing deposits")

	// get deposits
	deposits, err := d.sc.GetDeposits(ctx, blk, blk)
	if err != nil {
		return err
	}

	// process deposits
	for _, dep := range deposits {
		// get amount in big int
		amt, errb := big.NewInt(0).SetString(dep.Amount(), 10)
		if errb {
			d.log.Warn().Int64("height", blk).Str("tx", dep.Tx()).Str("amount", dep.Amount()).Msg("failed to parse amount")
			continue
		}
		if dep.Target() == "" {
			continue // TODO: make this check against the local node's address
		}
		err := d.ds.Deposit(dep.Tx(), dep.Caller(), amt, dep.Height())
		if err != nil {
			if err == store.ErrTxExists {
				d.log.Debug().Int64("height", blk).Str("tx", dep.Tx()).Str("amount", dep.Amount()).Msg("deposit already processed")
				continue
			} else {
				d.log.Warn().Err(err).Int64("height", blk).Str("tx", dep.Tx()).Str("amount", dep.Amount()).Err(err).Msg("failed to process deposit")
				continue
			}
		}
	}

	return d.ds.CommitBlock(blk)
}

func (d *deposits) processWithdrawals(ctx context.Context, blk int64) error {
	/*d.log.Debug().Int64("height", blk).Msg("processing withdrawals")

	// get withdrawals
	withdrawals, err := d.sc.GetWithdrawals(ctx, blk)
	if err != nil {
		return err
	}

	// process withdrawals
	for _, w := range withdrawals {
	}*/
	return nil
}

func (d *deposits) GetBalance(addr string) (*big.Int, error) {
	return d.ds.GetBalance(addr)
}

func (d *deposits) GetSpent(addr string) (*big.Int, error) {
	return d.ds.GetSpent(addr)
}

func (d *deposits) Spend(addr string, amt *big.Int) error {
	return d.ds.Spend(addr, amt)
}

func (d *deposits) Close() error {
	return d.ds.Close()
}

// sync syncs the deposits with the chain
func (d *deposits) Sync(ctx context.Context) error {
	return nil
}
