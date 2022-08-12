package deposits

import (
	"github.com/kwilteam/kwil-db/pkg/types"
	"math/big"
)

type KVStore interface {
	Get(key []byte) ([]byte, error)
	Set(key, val []byte) error
	RunGC()
}

const (
	DepositKey = "deposit"
)

type DepositStore struct {
	kvStore KVStore
	prefix  []byte
}

func New(kvStore KVStore) *DepositStore {
	go kvStore.RunGC()
	return &DepositStore{
		kvStore: kvStore,
		prefix:  []byte(DepositKey),
	}
}

// Deposit is a function that takes an amount and an address and adds the amount to the address's balance
func (ds *DepositStore) Deposit(amt *big.Int, addr string) error {
	// Get the key
	key := append(ds.prefix, []byte(addr)...)

	// Get the current amount\
	curAmt, err := ds.kvStore.Get(key) // I use this instead of GetBalance since GetBalance returns a bigint
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
	return ds.kvStore.Set(key, val)
}

// GetBalance returns the balance of the address
func (ds *DepositStore) GetBalance(addr string) (*big.Int, error) {
	key := append(ds.prefix, []byte(addr)...)
	val, err := ds.kvStore.Get(key)
	if err != nil {
		return nil, err
	}

	bal := Byte2BigInt(val)
	return bal, nil

	//return big.NewInt(int64(binary.LittleEndian.Uint64(val))), nil
}

func Byte2BigInt(b []byte) *big.Int {
	bi := big.NewInt(0)
	bi.SetBytes(b)
	return bi
}
