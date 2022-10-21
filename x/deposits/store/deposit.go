package store

import (
	"math/big"

	"kwil/x/cfgx"
	"kwil/x/deposits/store/kv"
	"kwil/x/logx"
)

var DEPOSITKEY = []byte("d")
var SPENTKEY = []byte("s")
var WITHDRAWALKEY = []byte("w")
var BLOCKKEY = []byte("b")
var LASTHEIGHT = []byte("l")

type depositStore struct {
	db kv.KVStore
}

type DepositStore interface {
	Deposit(string, string, *big.Int, int64) error
	GetBalance(string) (*big.Int, error)
	Spend(string, *big.Int) error
	CommitBlock(int64) error
	Close() error
	GetLastHeight() (int64, error)
	SetLastHeight(int64) error
	GetSpent(string) (*big.Int, error)
}

func New(conf cfgx.Config, l logx.Logger) (*depositStore, error) {

	db, err := kv.New(l, conf.String("kv.path"))
	if err != nil {
		return nil, err
	}

	return &depositStore{
		db: db,
	}, nil
}

func (ds *depositStore) Deposit(txid, addr string, amt *big.Int, h int64) error {
	// user's funding amount key
	key := append(DEPOSITKEY, []byte(addr)...)

	// tx key for idempotency.  We can delete all of these by block height later
	txKey := append(append(BLOCKKEY, int64ToBytes(h)...), []byte(txid)...)

	// get the current amount
	curAmt, err := ds.db.Get(key) // I use this instead of GetBalance since GetBalance returns a bigint
	if err != nil {
		if err == ErrNotFound { // If it could not find a value, then set the total to 0
			curAmt = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			return err
		}
	}

	// current amt in bigint
	a := new(big.Int).SetBytes(curAmt)

	// add the new amount to the current amount
	a.Add(a, amt)

	val := a.Bytes()

	// check to make sure the tx key does not exist yet.
	// if it does, then we have already begun processing this tx and the server crashed

	exists, err := ds.db.Exists(txKey)
	if err != nil {
		return err
	}

	if exists {
		return ErrTxExists
	}

	// now create transaction for the new amount
	txn := ds.db.NewTransaction(true)
	defer txn.Discard()
	err = txn.Set(key, val)
	if err != nil {
		return err
	}

	// now create transaction for the txid
	err = txn.Set(txKey, []byte{})
	if err != nil {
		return err
	}

	return txn.Commit()
}

// Spend moves the amount from the deposit to the spent bucket in the same transaction
func (ds *depositStore) Spend(addr string, amt *big.Int) error {
	// user's funding amount key
	spendKey := append(SPENTKEY, []byte(addr)...)

	a, err := ds.GetBalance(addr)
	if err != nil {
		return err
	}

	// subtract the new amount to the current amount
	a.Sub(a, amt)

	// check to make sure the amount is not negative
	if a.Cmp(big.NewInt(0)) == -1 {
		return ErrInsufficientFunds
	}

	newBal := a.Bytes() // new balance to be set in deposit bucket

	// get the current amount in the spent bucket
	curSpendAmt, err := ds.db.Get(spendKey)
	if err != nil {
		if err == ErrNotFound { // If it could not find a value, then set the total to 0
			curSpendAmt = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			return err
		}
	}

	// current amt in bigint
	spendAmt := new(big.Int).SetBytes(curSpendAmt)

	// add the new amount to the current amount
	spendAmt.Add(spendAmt, amt)

	newSpendAmt := spendAmt.Bytes()

	txn := ds.db.NewTransaction(true)
	defer txn.Discard()

	// set the new balance in the deposit bucket
	depKey := append(DEPOSITKEY, []byte(addr)...)
	err = txn.Set(depKey, newBal)
	if err != nil {
		return err
	}
	err = txn.Set(spendKey, newSpendAmt)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (ds *depositStore) GetBalance(addr string) (*big.Int, error) {
	key := append(DEPOSITKEY, []byte(addr)...)
	val, err := ds.db.Get(key)
	if err != nil {
		// if it could not find a value, then set the total to 0
		if err == kv.ErrNotFound {
			return big.NewInt(0), nil
		}
		return nil, err
	}

	return byteToBigInt(val), nil
}

// CommitBlock deletes all transactions with the block prefix.
// It will also increase the height by 1, in the same db transaction.
func (ds *depositStore) CommitBlock(h int64) error {

	key := append(BLOCKKEY, int64ToBytes(h)...)
	txn := ds.db.NewTransaction(true)
	defer txn.Discard()

	// get all keys with the block prefix
	keys, err := ds.db.Keys(key)
	if err != nil {
		return err
	}

	// delete all keys with the block prefix
	for _, k := range keys {
		err = txn.Delete(k)
		if err != nil {
			return err
		}
	}

	// increase the height by 1
	err = txn.Set(LASTHEIGHT, int64ToBytes(h+1))
	if err != nil {
		return err
	}

	return txn.Commit()
}

// SetLastHeight sets the last height to the given height.  It will commit the block at the given height
func (ds *depositStore) SetLastHeight(h int64) error {
	// get current height
	curH, err := ds.GetLastHeight()
	if err != nil {
		return err
	}

	// commit last height
	err = ds.CommitBlock(curH)
	if err != nil {
		return err
	}

	return ds.db.Set(LASTHEIGHT, int64ToBytes(h))
}

func (ds *depositStore) GetLastHeight() (int64, error) {
	val, err := ds.db.Get(LASTHEIGHT)

	if err != nil {
		if err == kv.ErrNotFound {
			val = int64ToBytes(0)
		} else {
			return 0, err
		}
	}

	return bytesToInt64(val), nil
}

func (ds *depositStore) Close() error {
	return ds.db.Close()
}

func (ds *depositStore) GetSpent(addr string) (*big.Int, error) {
	key := append(SPENTKEY, []byte(addr)...)
	val, err := ds.db.Get(key)
	if err != nil {
		// if it could not find a value, then set the total to 0
		if err == kv.ErrNotFound {
			return big.NewInt(0), nil
		}
		return nil, err
	}

	return byteToBigInt(val), nil
}
