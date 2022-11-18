package deposits

import (
	"context"
	"fmt"
	"kwil/x/async"
	"kwil/x/lease"
	"math/big"

	"kwil/x/cfgx"
	kc "kwil/x/crypto"
	"kwil/x/deposits/events"
	"kwil/x/deposits/store/sql"
	"kwil/x/deposits/types"
	"kwil/x/logx"
)

type Deposits interface {
	Listen(context.Context) error
	GetBalance(string) (*big.Int, error)
	GetSpent(string) (*big.Int, error)
	GetBalanceAndSpent(string) (string, string, error)
	Spend(address string, amount string) error
	Withdraw(context.Context, string, string) (*types.PendingWithdrawal, error)
	Close() error
	GetWithdrawalsForWallet(string) ([]*types.PendingWithdrawal, error)
	Address() string
}

type deposits struct {
	run  bool
	log  logx.SugaredLogger
	conf cfgx.Config
	ef   events.EventFeed
	sc   types.Contract
	lh   int64
	we   int64
	sql  sql.SQLStore
	acc  kc.Account
	addr string
}

func New(c cfgx.Config, l logx.Logger, acc kc.Account) (*deposits, error) {
	run, err := c.GetBool("run", false)
	if err != nil {
		return nil, fmt.Errorf("failed to get run from config. %w", err)
	}
	if !run {
		l.Sugar().Infof("deposits disabled")
		return &deposits{
			run:  false,
			addr: "0x0000000000000000000000000000000000000000",
		}, nil
	}

	we, err := c.GetInt64("withdrawal_expiration", 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal_expiration from config. %w", err)
	}

	pgConf, err := sql.NewConfig(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get pg config. %w", err)
	}

	pg, err := sql.New(pgConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql store. %w", err)
	}

	lb, err := pg.GetHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block from db. %w", err)
	}

	l.Sugar().Infof("last block height: %d", lb)

	ef, err := events.New(c, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize event feed. %w", err)
	}

	return &deposits{
		run:  run,
		log:  l.Sugar(),
		conf: c,
		ef:   ef,
		sc:   ef.Contract(),
		lh:   lb,
		we:   we,
		sql:  pg,
		acc:  acc,
		addr: acc.GetAddress().String(),
	}, nil
}

func (d *deposits) Listen(ctx context.Context) error {
	if !d.run {
		<-ctx.Done()
		return nil
	}

	agent, err := d.sql.CreateLeaseAgent("deposits_listener")
	if err != nil {
		return err
	}

	action := async.NewActionAsync()
	err = agent.Subscribe(ctx, "deposits_lock", lease.Subscriber{
		OnAcquired: func(l lease.Lease) {
			d.listen_safe(ctx, l.OnRevoked(), action)
		},
		OnFatalError: func(err error) {
			action.Fail(err)
		},
	})

	if err != nil {
		return err
	}

	<-action.DoneCh()

	return action.GetError()
}

func (d *deposits) listen_safe(ctx context.Context, stop <-chan struct{}, action async.Action) {
	// sync
	err := d.Sync(ctx)
	if err != nil {
		action.Fail(err)
		return
	}

	blks, errs, err := d.ef.Listen(ctx, d.lh)
	if err != nil {
		action.Fail(err)
		return
	}

	defer d.sql.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return // no longer the leader
		case err := <-errs:
			// TODO: do we start this back up or propagate the failure at all?
			d.log.Warnf("error from event feed: %v", err)
			return
		case blk := <-blks:
			err := d.processBlock(ctx, blk)
			if err != nil {
				// TODO: do we start this back up or propagate the failure at all?
				d.log.Warnf("failed to process block %d. %v", blk, err)
				return
			}
		}
	}
}

var ErrDepositsNotRunning = fmt.Errorf("deposits not running")

func (d *deposits) processBlock(ctx context.Context, blk int64) error {
	d.log.Infof("processing block %d", blk)
	// get deposits for the block
	deposits, err := d.sc.GetDeposits(ctx, blk, blk, d.addr)
	if err != nil {
		return fmt.Errorf("failed to get deposits for block %d. %w", blk, err)
	}

	for _, dep := range deposits {
		ttx := dep.Tx[2:]
		err = d.sql.Deposit(ttx, dep.Caller, dep.Amount, dep.Height)
		if err != nil {
			d.log.Errorf("failed to deposit %s. %v", dep.Tx, err)
			continue
		}
	}

	// get withdrawals for the block
	withdrawals, err := d.sc.GetWithdrawals(ctx, blk, blk, d.addr)
	if err != nil {
		return fmt.Errorf("failed to get withdrawals for block %d. %w", blk, err)
	}

	for _, wd := range withdrawals {
		exists, err := d.sql.FinishWithdrawal(wd.Cid)
		if err != nil {
			d.log.Errorf("failed to finish withdrawal %s. %v", wd.Cid, err)
			continue
		}

		if !exists {
			d.log.Warnf("withdrawal %s does not exist", wd.Cid)
		}
	}

	d.sql.CommitHeight(blk)

	return nil
}

func (d *deposits) GetBalance(addr string) (*big.Int, error) {
	if !d.run {
		return nil, ErrDepositsNotRunning
	}

	return d.sql.GetBalance(addr)
}

func (d *deposits) GetSpent(addr string) (*big.Int, error) {
	if !d.run {
		return nil, ErrDepositsNotRunning
	}

	return d.sql.GetSpent(addr)
}

// Spend will try to spend the amount from the address.
// If the addr does not have enough, it will return ErrInsufficientFunds
func (d *deposits) Spend(addr string, amt string) error {
	if !d.run {
		return ErrDepositsNotRunning
	}

	return d.sql.Spend(addr, amt)
}

func (d *deposits) GetBalanceAndSpent(addr string) (string, string, error) {
	if !d.run {
		return "", "", ErrDepositsNotRunning
	}

	return d.sql.GetBalanceAndSpent(addr)
}

func (d *deposits) Close() error {
	if d.sql == nil {
		return nil
	}

	return d.sql.Close()
}

func (d *deposits) Address() string {
	return d.addr
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
	if !d.run {
		return nil
	}

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
			ttx := dep.Tx[2:]
			err = d.sql.Deposit(ttx, dep.Caller, dep.Amount, chunk[0])
			if err != nil {
				d.log.Errorf("failed to deposit %s. %v", dep.Tx, err)
				return err // we want to return here since there is a major error
			}
		}

		// we now need to commit deposits.  This is the first half of committing a block/chunk, but we do not want to expire our withdrawals for the chunk yet
		err = d.sql.CommitDeposits(chunk[0])
		if err != nil {
			d.log.Errorf("failed to commit deposits for chunk %d. %v", chunk[0], err)
			return err
		}

		wdrls, err := d.sc.GetWithdrawals(ctx, chunk[0], chunk[1], d.addr)
		if err != nil {
			return err
		}

		for _, wdrl := range wdrls {
			exists, err := d.sql.FinishWithdrawal(wdrl.Cid)
			if err != nil {
				return err
			}

			d.log.Infof("withdrawal %s exists: %t", wdrl.Cid, exists)

			if !exists { // if the withdrawal did not exist, then take the fee and amount, add them, and spend them from the wallets balance.  This is b/c the withdrawal was already processed on-chain.
				// TODO: This covers 99% of cases, but it not s ecure enough for a productiuon env with real money.

				// parse fee and amount to big int
				bgf, ok := new(big.Int).SetString(wdrl.Fee, 10)
				if !ok {
					return err // since this only runs on startup, we want this to return.  Ethereum event logs should be in a correct format, so an error is a problem with our tech
				}

				bga, ok := new(big.Int).SetString(wdrl.Amount, 10)
				if !ok {
					return err
				}

				// add them
				bga.Add(bga, bgf)

				// spend them
				err = d.sql.RemoveBalance(wdrl.Receiver, bga.String())
				if err != nil {
					return err
				}
			}

			// we must also update the synced balance and spent resulting from a successful withdrawal
		}
		// commit the chunk
		d.lh = chunk[1]
		d.log.Infof("committing chunk, range %d to %d", chunk[0], chunk[1])
		err = d.sql.Expire(chunk[0]) // the second half of commit block is expire
		if err != nil {
			return err // returning here since if this happens then we want this to crash
		}
		// set height to chunk[1]+1
		err = d.sql.SetHeight(d.lh)
		if err != nil {
			return err
		}

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
