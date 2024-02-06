//go:build pglive

package deposit_oracle

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	ctypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/events"
	"github.com/kwilteam/kwil-db/internal/sql"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func Test_UnpackEventLogs(t *testing.T) {
	contractAbi, err := abi.JSON(strings.NewReader(contractABIStr))
	require.NoError(t, err)

	eventSig := crypto.Keccak256Hash([]byte(depositEventSignature))
	invalidSig := crypto.Keccak256Hash([]byte("Depsoit(address,address,uint256)"))

	address := common.HexToAddress("0x80b2c9d7cbbf30a1b0fc8983c647d754c6525615")
	amount := big.NewInt(100)
	addressType, err := abi.NewType("address", "", nil)
	require.NoError(t, err)
	uint256Type, err := abi.NewType("uint256", "", nil)
	require.NoError(t, err)

	addrArg := abi.Argument{Name: "_from", Type: addressType}
	amountArg := abi.Argument{Name: "_amount", Type: uint256Type}
	args := abi.Arguments{addrArg, amountArg}
	invalidArgs := abi.Arguments{addrArg, amountArg, amountArg}

	validData, err := args.Pack(address, amount)
	require.NoError(t, err)

	invalidData, err := invalidArgs.Pack(address, amount, amount)
	require.NoError(t, err)

	contractAddres := common.HexToAddress("0xBcF7FFFD8B256Ec51a36782a52D0c34f6474D951")
	txHash := common.HexToHash("0x71b25cc6cfb0738e4eeca169bf49a26f1d3f918dd7ddc70cd2b9ae1e865e5d3b")
	blockHash := common.HexToHash("0x4eb64902935943ac7b748ec732559e5c610f0d1068ff15605c1d96adabd6c5f8")

	oracle := &EthDepositOracle{
		eventABI:             contractAbi,
		creditEventSignature: eventSig,
		logger:               log.NewStdOut(log.InfoLevel),
	}

	type testcase struct {
		name string
		log  types.Log
		fn   func(*testing.T, types.Log)
	}
	testcases := []testcase{
		{
			name: "Log without Data",
			log: types.Log{
				Address:     contractAddres,
				Data:        []byte{},
				TxHash:      txHash,
				BlockHash:   blockHash,
				BlockNumber: 1,
				TxIndex:     1,
				Index:       1,
			},
			fn: func(t *testing.T, log types.Log) {
				evt, err := oracle.convertLogToCreditEvent(log)
				fmt.Println(evt)
				require.Error(t, err)
				require.Nil(t, evt)
			},
		},
		{
			name: "Log with invalid event data and signature",
			log: types.Log{
				Address: contractAddres,
				Topics: []common.Hash{
					invalidSig,
				},
				Data:        invalidData,
				TxHash:      txHash,
				BlockHash:   blockHash,
				BlockNumber: 1,
				TxIndex:     1,
				Index:       1,
			},
			fn: func(t *testing.T, log types.Log) {
				evt, err := oracle.convertLogToCreditEvent(log)
				require.Error(t, err)
				require.Nil(t, evt)
			},
		},
		{
			name: "Log with valid event data",
			log: types.Log{
				Address: contractAddres,
				Topics: []common.Hash{
					eventSig,
				},
				Data:        validData,
				TxHash:      txHash,
				BlockHash:   blockHash,
				BlockNumber: 1,
				TxIndex:     1,
				Index:       1,
			},
			fn: func(t *testing.T, log types.Log) {
				evt, err := oracle.convertLogToCreditEvent(log)
				fmt.Println(evt)
				require.NoError(t, err)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(t, tc.log)
		})
	}
}

func Test_AddEvent(t *testing.T) {
	ctx := context.Background()

	// no MockEventStore (yet?)
	db, cleanUp, err := dbtest.NewTestPool(ctx, []string{`kwild_events`}) // ugh, unexported literal
	require.NoError(t, err)
	defer cleanUp()

	es, err := events.NewEventStore(ctx, db)
	require.NoError(t, err)

	vs := NewMockVoteStore()
	em := events.NewEventMgr(es, vs)

	oracle := &EthDepositOracle{
		logger:     log.NewStdOut(log.InfoLevel),
		eventstore: em,
	}

	type testcase struct {
		name string
		ac   *AccountCredit
		fn   func(*testing.T, *AccountCredit)
	}
	testcases := []testcase{
		{
			name: "nil account credit",
			ac:   nil,
			fn: func(t *testing.T, ac *AccountCredit) {
				err := oracle.addEvent(ctx, ac)
				require.NoError(t, err)

				evts, err := em.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, evts, 0)
			},
		},
		{
			name: "valid account credit",
			ac: &AccountCredit{
				Account:   "0x80b2c9d7cbbf30a1b0fc8983c647d754c6525615",
				Amount:    big.NewInt(100),
				TxHash:    "0x71b25cc6cfb0738e4eeca169bf49a26f1d3f918dd7ddc70cd2b9ae1e865e5d3b",
				BlockHash: "0x4eb64902935943ac7b748ec732559e5c610f0d1068ff15605c1d96adabd6c5f8",
			},
			fn: func(t *testing.T, ac *AccountCredit) {
				err := oracle.addEvent(ctx, ac)
				require.NoError(t, err)

				evts, err := em.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, evts, 1)
				var acR AccountCredit
				err = acR.UnmarshalBinary(evts[0].Body)
				require.NoError(t, err)
				require.Equal(t, ac.Account, acR.Account)
				require.Equal(t, ac.Amount, acR.Amount)
			},
		},
		{
			name: "Already processed account credit",
			ac: &AccountCredit{
				Account:   "0x80b2c9d7cbbf30a1b0fc8983c647d754c6525615",
				Amount:    big.NewInt(100),
				TxHash:    "0x71b25cc6cfb0738e4eeca169bf49a26f1d3f918dd7ddc70cd2b9ae1e865e5d3b",
				BlockHash: "0x4eb64902935943ac7b748ec732559e5c610f0d1068ff15605c1d96adabd6c5f8",
			},
			fn: func(t *testing.T, ac *AccountCredit) {
				err := oracle.addEvent(ctx, ac)
				require.NoError(t, err)

				evts, err := em.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, evts, 1)

				vs.processed[evts[0].ID()] = true

				// Add procssed event, should be ignored by the event manager
				err = oracle.addEvent(ctx, ac)
				require.NoError(t, err)

				evts, err = em.GetEvents(ctx)
				require.NoError(t, err)

				require.Len(t, evts, 1)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			tc.fn(t, tc.ac)
		})
	}
}

type MockEventStore struct {
	events map[ctypes.UUID]int
}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{
		events: make(map[ctypes.UUID]int),
	}
}

func (m *MockEventStore) Store(ctx context.Context, data []byte, eventType string) error {
	ve := &ctypes.VotableEvent{
		Body: data,
		Type: eventType,
	}
	id := ve.ID()

	if _, ok := m.events[id]; !ok {
		m.events[id] = 0
	}

	m.events[id] = m.events[id] + 1

	return nil
}

func (m *MockEventStore) KVStore() sql.KV {
	return nil
}

type mockVoteStore struct {
	processed map[ctypes.UUID]bool
}

func NewMockVoteStore() *mockVoteStore {
	return &mockVoteStore{
		processed: make(map[ctypes.UUID]bool),
	}
}

func (m *mockVoteStore) Processed(resolutionID ctypes.UUID) {
	m.processed[resolutionID] = true
}

func (m *mockVoteStore) IsProcessed(ctx context.Context, resolutionID ctypes.UUID) (bool, error) {
	return m.processed[resolutionID], nil
}
