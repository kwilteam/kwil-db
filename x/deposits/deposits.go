package deposits

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"kwil/x/cfgx"
	kc "kwil/x/crypto"
	ct "kwil/x/deposits/chainclient/types"
	"kwil/x/deposits/events"
	"kwil/x/deposits/store"
	"kwil/x/logx"
)

type Deposits interface {
	Listen(context.Context) error
	GetBalance(string) (*big.Int, error)
	GetSpent(string) (*big.Int, error)
	Spend(string, *big.Int) error
	Close() error
}

type deposits struct {
	log  logx.SugaredLogger
	conf cfgx.Config
	ef   events.EventFeed
	sc   ct.Contract
	lh   int64
	ds   store.DepositStore
	acc  kc.Account
	addr string
}

func New(c cfgx.Config, l logx.Logger, acc kc.Account) (*deposits, error) {

	ds, err := store.New(c, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize deposit store. %w", err)
	}

	lb, err := ds.GetLastHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block height. %w", err)
	}

	ef, err := events.New(c, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event feed. %w", err)
	}

	return &deposits{
		log:  l.Sugar(),
		conf: c,
		ef:   ef,
		sc:   ef.Contract(),
		lh:   lb,
		ds:   ds,
		acc:  acc,
		addr: acc.GetAddress().Hex(),
	}, nil
}

func (d *deposits) Listen(ctx context.Context) error {

	// sync
	err := d.Sync(ctx)
	if err != nil {
		return err
	}

	blks, errs, err := d.ef.Listen(ctx, d.lh)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				d.log.Warnf("error from event feed: %v", err)
				return
			case blk := <-blks:
				d.processBlock(ctx, blk)
			}
		}
	}()

	return nil
}

func (d *deposits) processBlock(ctx context.Context, blk int64) error {
	d.log.Infof("processing block %d", blk)
	wg := sync.WaitGroup{}
	wg.Add(2) // processing deposits and withdrawals simultaneously

	go func(context.Context, int64) {
		// process deposits
		err := d.processDeposits(ctx, blk)
		if err != nil {
			d.log.Warnf("failed to process deposits for block %d. %v", blk, err)
		}
		wg.Done()
	}(ctx, blk)

	go func(context.Context, int64) {
		// process withdrawals
		err := d.processWithdrawals(ctx, blk)
		if err != nil {
			d.log.Warnf("failed to process withdrawals for block %d. %v", blk, err)
		}
		wg.Done()
	}(ctx, blk)

	wg.Wait()
	d.lh = blk + 1
	return d.ds.CommitBlock(blk)
}

func (d *deposits) processDeposits(ctx context.Context, blk int64) error {

	// get deposits
	depos, err := d.sc.GetDeposits(ctx, blk, blk, d.addr)
	if err != nil {
		return err
	}

	// process deposits
	for _, dep := range depos {
		// get amount in big int
		amt, errb := big.NewInt(0).SetString(dep.Amount(), 10)
		if errb {
			d.log.Errorf("failed to convert amount to big int.  amt: %s | tx: %s | err: %v", dep.Amount(), dep.Tx(), errb)
			continue
		}
		if dep.Target() != d.addr { // only process deposits to this address
			continue
		}
		err := d.ds.Deposit(dep.Tx(), dep.Caller(), amt, dep.Height())
		if err != nil {
			if err == store.ErrTxExists {
				d.log.Debugf("deposit already processed. tx: %s", dep.Tx())
				continue
			} else {
				d.log.Errorf("failed to process deposit. tx: %s | err: %v", dep.Tx(), err)
				continue
			}
		}
	}

	return nil
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

/*
	Sync will sync the deposit store with the blockchain.

	It starts by getting the last confirmed block from the client chain.
	It then gets the last processed block height from the deposit store.

	It then splits these into chunks of blocks

	For each chunk it will get the deposits from the chain
	Withdrawals aren't necessary since validators (e.g. us) trigger them

	It will loop through each deposit and process it
	It will then commit the block height to the deposit store, auto incrementing the last processed block height
*/

// TODO: if on the last chunk the db crashes, the last chunk will get partially processed but not confirmed.
// since the last chunk is not a "full" chunk (e.g. having 100,000 blocks), the txKey generated before the crash
// will be different than the one after the crash.  This will cause the tx to be processed again.

// sync syncs the deposits with the chain
func (d *deposits) Sync(ctx context.Context) error {
	d.log.Info("syncing deposits...")
	lb, err := d.ef.GetLastConfirmedBlock(ctx)
	if err != nil {
		return err
	}

	d.log.Infof("syncing from block %d to %d", d.lh, lb)

	chunks := splitBlocks(d.lh, lb, d.conf.Int64("sync.chunk-size", 10000))

	for _, chunk := range chunks {
		// get deposits for the chunk
		deps, err := d.sc.GetDeposits(ctx, chunk[0], chunk[1], d.addr)
		if err != nil {
			return err
		}

		// we can now batch insert the deps
		// the height we use will be the last height in the chunk
		for _, dep := range deps {
			// get amount in big int
			amt, errb := big.NewInt(0).SetString(dep.Amount(), 10)
			if errb {
				d.log.Errorf("failed to convert amount to big int.  amt: %s | tx: %s | chunk-end: | err: %v", dep.Amount(), dep.Tx(), chunk[1], errb)
				continue
			}
			err := d.ds.Deposit(dep.Tx(), dep.Caller(), amt, chunk[1])
			if err != nil {
				d.log.Errorf("failed to process deposit. tx: %s | chunk-end: %d | err: %v", dep.Tx(), chunk[1], err)
				continue
			}
		}

		// commit the chunk
		d.log.Debugf("committing chunk, range %d to %d", chunk[0], chunk[1])
		d.lh = chunk[1] + 1
		err = d.ds.CommitBlock(chunk[1])
		if err != nil {
			d.log.Errorf("failed to commit chunk.  chunk-end: %d | err: %v", chunk[1], err)
			return err
		}
		// last height should now be n+1 of the last height in the chunk that was just processed
	}
	return nil
}

/*
split into chunks of n blocks

e.g. if we are at block 0 and the last block is 350,000 and chunk-size is 100,000,
we will process [0, 99999] [100000, 199999], [200000, 299999], [300000, 349999]

this technically means that the very last block won't be included, but as soon as the next block
gets received it will be recognized as being too high and be compensated for
*/
type chunk [2]int64

func splitBlocks(start, end, chunkSize int64) []chunk {
	var chunks []chunk
	for i := start; i < end; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > end {
			chunkEnd = end
		}
		chunks = append(chunks, chunk{i, chunkEnd - 1})
	}
	return chunks
}
