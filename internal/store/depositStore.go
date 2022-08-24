package store

/*
   Pretty much all of this is just an abstraction on top of basic kvstore functions
*/

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog/log"
	"math/big"
)

var DepositKey = []byte("deposit")
var BlockKey = []byte("block")
var HeightKey = []byte("lh") // stands for last height

type DepositStore struct {
	db     *BadgerDB
	prefix []byte
	wal    types.Wal
}

func NewDepositStore(conf *types.Config, wal types.Wal) (*DepositStore, error) {
	kvStore, err := New(conf)
	//kvStore.PrintAll()
	if err != nil {
		return nil, err
	}

	go kvStore.RunGC() // Garbage collection
	ds := &DepositStore{
		db:     kvStore,
		prefix: DepositKey,
		wal:    wal,
	}

	lh, err := ds.GetLastHeight()
	if err != nil {
		return nil, err
	}
	if lh.Cmp(big.NewInt(0)) == 0 {
		ds.SetLastHeight(lh)
	}

	return ds, nil
}

// Deposit is a function that takes an amount and an address and adds the amount to the address's balance
func (ds *DepositStore) Deposit(amt *big.Int, addr string, tx []byte, height *big.Int) error {
	// Get the key
	key := append(ds.prefix, []byte(addr)...)
	// Create a key based on the height and tx
	txKey := append(append(BlockKey, height.Bytes()...), tx...)

	// Get the current amount\
	curAmt, err := ds.db.Get(key) // I use this instead of GetBalance since GetBalance returns a bigint
	if err != nil {
		if err == types.ErrNotFound { // If it could not find a value, then set the total to 0
			curAmt = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			return err
		}
	}

	// Get the current amount as big int
	a := Byte2BigInt(curAmt)

	// Add
	amt.Add(amt, a)

	val := amt.Bytes()

	exists, err := ds.db.Exists(txKey)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Now we create a transaction to store both the balance and txKey
	txn := ds.db.NewTransaction(true)
	defer txn.Discard()
	err = txn.Set(key, val)
	if err != nil {
		return err
	}

	err = txn.Set(txKey, []byte{})
	if err != nil {
		return err
	}

	return txn.Commit()
}

// CommitBlock deletes all the transactions in the block
func (ds *DepositStore) CommitBlock(height *big.Int) error {

	pref := append(BlockKey, height.Bytes()...) // bht is the prefix for block height
	err := ds.db.DeleteByPrefix(pref)
	if err != nil {
		return err
	}

	// TODO: You should be able to just update the block height as shown:
	//ds.SetLastHeight(height)

	// I am not doing this right now to ensure that this behaves as expected.
	// For production, the above should be used.

	// Now update the last height
	prevH, err := ds.GetLastHeight()
	if err != nil {
		return err
	}
	curH := prevH.Add(prevH, big.NewInt(1))
	if curH.Cmp(height) == 0 {
		ds.SetLastHeight(curH)
	} else {
		log.Fatal().Str("received from blockchain:", height.String()).Str("incremented", curH.String()).Msgf("expected height not received")
	}

	return nil
}

// GetBalance returns the balance of the address
func (ds *DepositStore) GetBalance(addr string) (*big.Int, error) {
	key := append(ds.prefix, []byte(addr)...)
	val, err := ds.db.Get(key)
	if err != nil {
		return nil, err
	}

	bal := Byte2BigInt(val)
	return bal, nil
}

func Byte2BigInt(b []byte) *big.Int {
	bi := big.NewInt(0)
	bi.SetBytes(b)
	return bi
}

func (ds *DepositStore) Close() {
	ds.db.Close()
}

// Prints all balances of all wallets
func (ds *DepositStore) PrintAllBalances() {
	keys, vals, err := ds.db.GetAllByPrefix(DepositKey)
	if err != nil {
		panic(err) // panic here since this is a test function
	}

	// Iterate through all the keys and print the values
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		key = StripPrefix(key, DepositKey)
		val := vals[i]
		bal := Byte2BigInt(val)
		fmt.Printf("Wallet: %s | Balance: %s\n", key, bal)
	}
}

func StripPrefix(key []byte, prefix []byte) []byte {
	return key[len(prefix):]
}

func (ds *DepositStore) GetLastHeight() (*big.Int, error) {
	val, err := ds.db.Get(HeightKey)
	if err != nil {
		if err == types.ErrNotFound { // If it could not find a value, then set the total to 0
			val = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			return Byte2BigInt(val), err
		}
	}
	return Byte2BigInt(val), nil
}

func (ds *DepositStore) SetLastHeight(height *big.Int) error {
	return ds.db.Set(HeightKey, height.Bytes())
}

func (ds *DepositStore) PrintCurrentHeight() {
	height, err := ds.GetLastHeight()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Last synced height: %s\n", height.String())
}
