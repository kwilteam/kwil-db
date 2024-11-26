//go:build pglive

package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/pg"
	pgtest "github.com/kwilteam/kwil-db/node/pg/test"
	"github.com/kwilteam/kwil-db/node/store/memstore"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

func TestDualNodeMocknet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	mn := mock.New()

	pk1, h1, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}
	bs1 := memstore.NewMemBS()
	mp1 := mempool.New()

	db1 := initDB(t, "5432", "kwil_test_db")

	pk2, h2, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}
	bs2 := memstore.NewMemBS()
	mp2 := mempool.New()
	db2 := initDB(t, "5432", "kwil_test_db2")

	root1 := t.TempDir()
	root2 := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	t.Cleanup(func() {
		fmt.Println("cleanup- cancel")
		cancel()
		fmt.Println("cleanup- wait")
		wg.Wait()

		fmt.Println("cleanup- db")
		cleanupDB(db1)
		cleanupDB(db2)
		fmt.Print("cleanup- done")
	})

	privKeys, _ := newGenesis(t, [][]byte{pk1, pk2})

	valSet := make(map[string]ktypes.Validator)
	for _, priv := range privKeys {
		valSet[hex.EncodeToString(priv.Public().Bytes())] = ktypes.Validator{
			PubKey: priv.Public().Bytes(),
			Power:  1,
		}
	}
	valSetList := make([]*ktypes.Validator, 0, len(valSet))
	for _, v := range valSet {
		valSetList = append(valSetList, &v)
	}

	ceCfg1 := &consensus.Config{
		PrivateKey:     privKeys[0],
		ValidatorSet:   valSet,
		Leader:         privKeys[0].Public(),
		Mempool:        mp1,
		BlockStore:     bs1,
		Accounts:       &mockAccounts{},
		ValidatorStore: newValidatorStore(valSetList),
		TxApp:          newDummyTxApp(valSetList),
		Logger:         log.New(log.WithName("CE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured)),
		ProposeTimeout: 1 * time.Second,
		DB:             db1,
	}
	ce1 := consensus.New(ceCfg1)
	defaultConfigSet := config.DefaultConfig()

	log1 := log.New(log.WithName("NODE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg1 := &Config{
		RootDir:    root1,
		PrivKey:    privKeys[0],
		Logger:     log1,
		P2P:        &defaultConfigSet.P2P,
		Mempool:    mp1,
		BlockStore: bs1,
		Consensus:  ce1,
	}
	node1, err := NewNode(cfg1, WithHost(h1))
	if err != nil {
		t.Fatalf("Failed to create Node 1: %v", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node1.Dir())
		node1.Start(ctx)
	}()

	// time.Sleep(200 * time.Millisecond) // !!!! apparently, needs this if block store does not have latency, so there is a race condition somewhere in CE
	time.Sleep(20 * time.Millisecond)

	ceCfg2 := &consensus.Config{
		PrivateKey:     privKeys[1],
		ValidatorSet:   valSet,
		Leader:         privKeys[0].Public(),
		Mempool:        mp2,
		BlockStore:     bs2,
		Accounts:       &mockAccounts{},
		ValidatorStore: newValidatorStore(valSetList),
		TxApp:          newDummyTxApp(valSetList),
		Logger:         log.New(log.WithName("CE2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured)),
		ProposeTimeout: 1 * time.Second,
		DB:             db2,
	}
	ce2 := consensus.New(ceCfg2)

	log2 := log.New(log.WithName("NODE2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg2 := &Config{
		RootDir:    root2,
		PrivKey:    privKeys[1],
		Logger:     log2,
		P2P:        &defaultConfigSet.P2P,
		Mempool:    mp2,
		BlockStore: bs2,
		Consensus:  ce2,
	}
	node2, err := NewNode(cfg2, WithHost(h2))
	if err != nil {
		t.Fatalf("Failed to create Node 2: %v", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node2.Dir())
		node2.Start(ctx)
	}()

	// Link and connect the hosts
	if err := mn.LinkAll(); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}

	// n1 := mn.Net(h1.ID())
	// links := mn.LinksBetweenPeers(h1.ID(), h2.ID())
	// ln := links[0]
	// net := ln.Networks()[0]
	// peers := net.Peers()
	// t.Log(peers)

	// run for a bit, checks stuff, do tests, like ensure blocks mine (TODO)...
	time.Sleep(4 * time.Second)

	tx, err := db1.BeginReadTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	h, _, _, err := meta.GetChainState(ctx, tx)
	require.NoError(t, err)

	require.GreaterOrEqual(t, h, int64(1))

	cancel()
	wg.Wait()

}

func initDB(t *testing.T, port, dbName string) *pg.DB {
	cfg := &config.DBConfig{
		Host:   "127.0.0.1",
		Port:   port,
		User:   "kwild",
		Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
		DBName: dbName,
	}
	db, err := pgtest.NewTestDBWithCfg(t, cfg)
	require.NoError(t, err)
	ctx := context.Background()

	prepTx, err := db.BeginPreparedTx(ctx)
	require.NoError(t, err)

	err = meta.InitializeMetaStore(ctx, prepTx)
	assert.NoError(t, err)

	assert.NoError(t, prepTx.Commit(ctx))
	return db
}

func cleanupDB(db *pg.DB) {
	ctx := context.Background()
	db.AutoCommit(true)
	defer db.AutoCommit(false)
	defer db.Close()
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_chain CASCADE;`)
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_internal CASCADE;`)
}

type dummyTxApp struct {
	vals []*ktypes.Validator
}

func newDummyTxApp(valset []*ktypes.Validator) *dummyTxApp {
	return &dummyTxApp{
		vals: valset,
	}
}
func (d *dummyTxApp) Begin(ctx context.Context, height int64) error {
	return nil
}

func (d *dummyTxApp) Execute(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse {
	return &txapp.TxResponse{}
}

func (d *dummyTxApp) Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) ([]*ktypes.Validator, error) {
	return d.vals, nil
}

func (d *dummyTxApp) Commit() error {
	return nil
}

type validatorStore struct {
	valSet []*ktypes.Validator
}

func newValidatorStore(valSet []*ktypes.Validator) *validatorStore {
	return &validatorStore{
		valSet: valSet,
	}
}

func (v *validatorStore) GetValidators() []*ktypes.Validator {
	return v.valSet
}

func (v *validatorStore) ValidatorUpdates() map[string]*ktypes.Validator {
	return nil
}

type mockAccounts struct{}

func (m *mockAccounts) Updates() []*ktypes.Account {
	return nil
}
