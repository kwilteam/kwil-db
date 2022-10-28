package deposits

import (
	"context"
	"fmt"
	"math/big"

	"kwil/x/cfgx"
	kc "kwil/x/crypto"
	"kwil/x/deposits/events"
	"kwil/x/deposits/processor"
	"kwil/x/deposits/store"
	ct "kwil/x/deposits/types"
	"kwil/x/logx"
	"kwil/x/svcx/messaging/mx"
	"kwil/x/svcx/wallet"
)

type Deposits interface {
	Listen(context.Context) error
	GetBalance(string) (*big.Int, error)
	GetSpent(string) (*big.Int, error)
	Spend(string, *big.Int) error
	Withdraw(string, *big.Int) error
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
	svc  wallet.RequestService
	prsr wallet.RequestProcessor
}

func New(c cfgx.Config, l logx.Logger, acc kc.Account, svc wallet.RequestService) (*deposits, error) {

	ds, err := store.New(c, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize deposit store. %w", err)
	}

	// TODO: Get last height from kafka instead of db
	lb, err := ds.GetLastHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block height. %w", err)
	}

	l.Sugar().Infof("last block height: %d", lb)

	ef, err := events.New(c, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event feed. %w", err)
	}

	pr := processor.NewProcessor(l)

	prsr, err := wallet.NewRequestProcessor(c, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize request processor. %w", err)
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
		svc:  svc,
		prsr: prsr,
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

	go func(*deposits) {
		defer d.ds.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				d.log.Warnf("error from event feed: %v", err)
				return
			case blk := <-blks:
				err := d.processBlock(ctx, blk)
				if err != nil {
					d.log.Warnf("failed to process block %d. %v", blk, err)
					return
				}
			}
		}
	}(d)

	return nil
}

func (d *deposits) processBlock(ctx context.Context, blk int64) error {
	d.log.Infof("processing block %d", blk)
	// get deposits for the block
	deposits, err := d.sc.GetDeposits(ctx, blk, blk, d.addr)
	if err != nil {
		return fmt.Errorf("failed to get deposits for block %d. %w", blk, err)
	}

	for _, dep := range deposits {
		bts, err := dep.Serialize()
		if err != nil {
			d.log.Warnf("failed to serialize deposit. %v", err)
			continue
		}

		d.svc.SubmitAsync(ctx, &mx.RawMessage{Key: []byte(dep.Caller()), Value: bts})
	}

	// get withdrawals for the block
	withdrawals, err := d.sc.GetWithdrawals(ctx, blk, blk, d.addr)
	if err != nil {
		return fmt.Errorf("failed to get withdrawals for block %d. %w", blk, err)
	}

	for _, wd := range withdrawals {
		bts, err := wd.Serialize()
		if err != nil {
			d.log.Warnf("failed to serialize withdrawal confirmation. %v", err)
			continue
		}

		d.svc.SubmitAsync(ctx, &mx.RawMessage{Key: []byte(wd.Caller()), Value: bts})
	}

	// TODO: Send end block to all partitions
	d.svc.SubmitAsync(ctx, &mx.RawMessage{Key: []byte("block"), Value: []byte(fmt.Sprintf("%d", blk))})

	return nil
}

/*

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
	return d.ds.CommitBlock(blk, d.lh)
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
		amt, ok := big.NewInt(0).SetString(dep.Amount(), 10)
		if !ok {
			d.log.Errorf("failed to convert amount to big int.  amt: %s | tx: %s | ok: %v", dep.Amount(), dep.Tx(), ok)
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

		d.log.Infof("processed deposit. tx: %s | caller: %s | amount: %s | height: %d", dep.Tx(), dep.Caller(), dep.Amount(), dep.Height())
	}

	return nil
}*/

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
	It will then commit the block height to the deposit store
	Chunks are identified and committed by the first block in the chunk
	The last block in the chunk needs to be auto incremented
*/

// sync syncs the deposits with the chain
func (d *deposits) Sync(ctx context.Context) error {
	lb, err := d.ef.GetLastConfirmedBlock(ctx)
	if err != nil {
		return err
	}

	d.log.Infof("syncing deposits from block %d to %d...", d.lh, lb)

	if d.lh == lb+1 {
		// already synced
		return nil
	}

	chunks := splitBlocks(d.lh, lb, d.conf.Int64("sync.chunk-size", 10000))

	d.log.Infof("syncing in %d chunks", len(chunks))
	for _, chunk := range chunks {
		// get deposits for the chunk
		deps, err := d.sc.GetDeposits(ctx, chunk[0], chunk[1], d.addr)
		if err != nil {
			return err
		}

		// we can now batch insert the deps
		// the height we use will be the last height in the chunk
		for _, dep := range deps {
			bts, err := dep.Serialize()
			if err != nil {
				d.log.Errorf("failed to serialize deposit.  amt: %s | tx: %s | chunk-end: | ok: %v", dep.Amount(), dep.Tx(), chunk[0])
				continue
			}

			d.svc.SubmitAsync(ctx, &mx.RawMessage{Key: []byte(dep.Caller()), Value: bts})
		}

		wdrls, err := d.sc.GetWithdrawals(ctx, chunk[0], chunk[1], d.addr)
		if err != nil {
			return err
		}

		for _, wdrl := range wdrls {
			bts, err := wdrl.Serialize()
			if err != nil {
				d.log.Errorf("failed to serialize withdrawal.  amt: %s | tx: %s | chunk-end: | ok: %v", wdrl.Amount(), wdrl.Tx(), chunk[0])
				continue
			}
			d.svc.SubmitAsync(ctx, &mx.RawMessage{Key: []byte(wdrl.Caller()), Value: bts})
		}

		// commit the chunk
		d.lh = chunk[1]
		d.log.Infof("committing chunk, range %d to %d", chunk[0], chunk[1])
		d.svc.SubmitAsync(ctx, &mx.RawMessage{Key: []byte("block"), Value: []byte(fmt.Sprintf("%d", chunk[1]))})

	}

	d.log.Infof("synced deposits to block %d", d.lh)
	return nil
}

/*
split into chunks of n blocks

e.g. if we are at block 0 and the last block is 350,000 and chunk-size is 100,000,
we will process [0, 99999] [100000, 199999], [200000, 299999], [300000, 349999]

the last chunk should have an additional block added to it
*/
type chunk [2]int64

func splitBlocks(start, end, chunkSize int64) []chunk {
	if start == end {
		return []chunk{{start, start}}
	}
	var chunks []chunk
	for i := start; i < end; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > end {
			chunkEnd = end
		}
		chunks = append(chunks, chunk{i, chunkEnd - 1})
	}

	if chunks[len(chunks)-1][1] != end {
		chunks[len(chunks)-1][1] = end
	}
	return chunks
}
