//go:build pglive

package node

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/accounts"
	blockprocessor "github.com/kwilteam/kwil-db/node/block_processor"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/migrations"
	"github.com/kwilteam/kwil-db/node/pg"
	pgtest "github.com/kwilteam/kwil-db/node/pg/test"
	"github.com/kwilteam/kwil-db/node/store/memstore"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/kwilteam/kwil-db/node/voting"
)

func TestMain(m *testing.M) {
	pg.UseLogger(log.New(log.WithName("DBS"), log.WithFormat(log.FormatUnstructured)))
	m.Run()
}

func TestSingleNodeMocknet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	mn := mock.New()

	pk1, h1 := newTestHost(t, mn)
	bs1 := memstore.NewMemBS()
	mp1 := mempool.New()

	db1 := initDB(t, "5432", "kwil_test_db")

	root1 := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	t.Cleanup(func() {
		cancel()
		wg.Wait()
		cleanupDB(db1)
	})

	privKeys, _ := newGenesis(t, [][]byte{pk1})

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

	ss := newSnapshotStore()

	_, vsReal, err := voting.NewResolutionStore(ctx, db1)
	require.NoError(t, err)

	genCfg := config.DefaultGenesisConfig()
	genCfg.Leader = privKeys[0].Public().Bytes()
	genCfg.Validators = valSetList

	k, err := crypto.UnmarshalSecp256k1PrivateKey(pk1)
	require.NoError(t, err)

	signer1 := &auth.EthPersonalSigner{Key: *k}

	es := &mockEventStore{}
	accounts := &mockAccounts{}
	mparams := config.MigrationParams{
		StartHeight: 0, EndHeight: 0,
	}

	migrator, err := migrations.SetupMigrator(ctx, db1, newSnapshotStore(), accounts, filepath.Join(root1, "migrations"), mparams, vsReal, log.New(log.WithName("MIGRATOR")))
	require.NoError(t, err)

	bpl := log.New(log.WithName("BP1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	bp, err := blockprocessor.NewBlockProcessor(ctx, db1, newDummyTxApp(valSetList), &mockAccounts{}, vsReal, ss, es, migrator, bs1, genCfg, signer1, bpl)
	require.NoError(t, err)

	ceCfg1 := &consensus.Config{
		PrivateKey:            privKeys[0],
		ValidatorSet:          valSet,
		Leader:                privKeys[0].Public(),
		Mempool:               mp1,
		BlockStore:            bs1,
		BlockProcessor:        bp,
		Logger:                log.New(log.WithName("CE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured)),
		ProposeTimeout:        1 * time.Second,
		BlockProposalInterval: 1 * time.Second,
		BlockAnnInterval:      3 * time.Second,
		DB:                    db1,
	}
	ce1 := consensus.New(ceCfg1)
	defaultConfigSet := config.DefaultConfig()
	log1 := log.New(log.WithName("NODE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg1 := &Config{
		ChainID:     genCfg.ChainID,
		RootDir:     root1,
		PrivKey:     privKeys[0],
		Logger:      log1,
		P2P:         &defaultConfigSet.P2P,
		Mempool:     mp1,
		BlockStore:  bs1,
		Consensus:   ce1,
		Snapshotter: ss,
		DBConfig:    &defaultConfigSet.DB,
		Statesync:   &defaultConfigSet.StateSync,
		BlockProc:   &dummyBP{vals: valSetList},
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

	time.Sleep(20 * time.Millisecond)

	reachHeight := int64(2)
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		stat, err := node1.Status(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(c, stat.Sync.BestBlockHeight, reachHeight)
	}, 10*time.Second, 250*time.Millisecond)

	cancel()
	wg.Wait()
}

func TestDualNodeMocknet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	mn := mock.New()

	pk1, h1 := newTestHost(t, mn)
	bs1 := memstore.NewMemBS()
	mp1 := mempool.New()

	db1 := initDB(t, "5432", "kwil_test_db")
	func() {
		ctx := context.Background()
		_, err := db1.Pool().Execute(ctx, `DROP DATABASE IF EXISTS kwil_test_db2;`)
		require.NoError(t, err)
		_, err = db1.Pool().Execute(ctx, `CREATE DATABASE kwil_test_db2 OWNER kwild;`)
		require.NoError(t, err)
	}()

	pk2, h2 := newTestHost(t, mn)
	bs2 := memstore.NewMemBS()
	mp2 := mempool.New()
	db2 := initDB(t, "5432", "kwil_test_db2") // NOTE: using the same postgres host is a little wild

	root1 := t.TempDir()
	root2 := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	t.Cleanup(func() {
		cancel()
		wg.Wait()
		cleanupDB(db1)
		cleanupDB(db2)
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
	ss := newSnapshotStore()

	genCfg := config.DefaultGenesisConfig()
	genCfg.Leader = privKeys[0].Public().Bytes()
	genCfg.Validators = valSetList

	// _, vsReal, err := voting.NewResolutionStore(ctx, db1)

	k, err := crypto.UnmarshalSecp256k1PrivateKey(pk1)
	require.NoError(t, err)

	signer1 := &auth.EthPersonalSigner{Key: *k}
	es1 := &mockEventStore{}
	accounts1 := &mockAccounts{}
	mparams := config.MigrationParams{
		StartHeight: 0, EndHeight: 0,
	}

	_, vstore1, err := voting.NewResolutionStore(ctx, db1)
	require.NoError(t, err)

	migrator, err := migrations.SetupMigrator(ctx, db1, newSnapshotStore(), accounts1, filepath.Join(root1, "migrations"), mparams, vstore1, log.New(log.WithName("MIGRATOR")))
	require.NoError(t, err)

	bpl1 := log.New(log.WithName("BP1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	bp1, err := blockprocessor.NewBlockProcessor(ctx, db1, newDummyTxApp(valSetList), accounts1, vstore1, ss, es1, migrator, bs1, genCfg, signer1, bpl1)
	require.NoError(t, err)

	ceCfg1 := &consensus.Config{
		PrivateKey:            privKeys[0],
		ValidatorSet:          valSet,
		Leader:                privKeys[0].Public(),
		Mempool:               mp1,
		BlockStore:            bs1,
		Logger:                log.New(log.WithName("CE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured)),
		ProposeTimeout:        1 * time.Second,
		BlockProposalInterval: 1 * time.Second,
		BlockAnnInterval:      3 * time.Second,
		DB:                    db1,
		BlockProcessor:        bp1,
	}
	ce1 := consensus.New(ceCfg1)
	defaultConfigSet := config.DefaultConfig()
	log1 := log.New(log.WithName("NODE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg1 := &Config{
		ChainID:     genCfg.ChainID,
		RootDir:     root1,
		PrivKey:     privKeys[0],
		Logger:      log1,
		P2P:         &defaultConfigSet.P2P,
		Mempool:     mp1,
		BlockStore:  bs1,
		Consensus:   ce1,
		Snapshotter: ss,
		DBConfig:    &defaultConfigSet.DB,
		Statesync:   &defaultConfigSet.StateSync,
		BlockProc:   &dummyBP{vals: valSetList},
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

	time.Sleep(20 * time.Millisecond)

	k2, err := crypto.UnmarshalSecp256k1PrivateKey(pk2)
	require.NoError(t, err)

	signer2 := &auth.EthPersonalSigner{Key: *k2}
	es2 := &mockEventStore{}
	accounts2 := &mockAccounts{}
	_, vstore2, err := voting.NewResolutionStore(ctx, db2)
	require.NoError(t, err)

	migrator2, err := migrations.SetupMigrator(ctx, db2, newSnapshotStore(), accounts2, filepath.Join(root2, "migrations"), mparams, vstore2, log.New(log.WithName("MIGRATOR")))
	require.NoError(t, err)

	bpl2 := log.New(log.WithName("BP2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	bp2, err := blockprocessor.NewBlockProcessor(ctx, db2, newDummyTxApp(valSetList), accounts2, vstore2, ss, es2, migrator2, bs2, genCfg, signer2, bpl2)
	require.NoError(t, err)

	ceCfg2 := &consensus.Config{
		PrivateKey:            privKeys[1],
		ValidatorSet:          valSet,
		Leader:                privKeys[0].Public(),
		Mempool:               mp2,
		BlockStore:            bs2,
		BlockProcessor:        bp2,
		Logger:                log.New(log.WithName("CE2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured)),
		ProposeTimeout:        1 * time.Second,
		BlockProposalInterval: 1 * time.Second,
		BlockAnnInterval:      3 * time.Second,
		DB:                    db2,
	}
	ce2 := consensus.New(ceCfg2)

	log2 := log.New(log.WithName("NODE2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg2 := &Config{
		RootDir:     root2,
		PrivKey:     privKeys[1],
		Logger:      log2,
		P2P:         &defaultConfigSet.P2P,
		Mempool:     mp2,
		BlockStore:  bs2,
		Consensus:   ce2,
		Snapshotter: ss,
		DBConfig:    &defaultConfigSet.DB,
		Statesync:   &defaultConfigSet.StateSync,
		BlockProc:   &dummyBP{vals: valSetList},
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

	reachHeight := int64(2)
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		stat, err := node1.Status(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(c, stat.Sync.BestBlockHeight, reachHeight)
	}, 10*time.Second, 250*time.Millisecond)

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

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	err = meta.InitializeMetaStore(ctx, tx)
	assert.NoError(t, err)

	assert.NoError(t, tx.Commit(ctx))
	return db
}

func cleanupDB(db *pg.DB) {
	defer db.Close()
	db.AutoCommit(true)
	ctx := context.Background()
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_chain CASCADE;`)
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_voting CASCADE;`)
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_events CASCADE;`)
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_migrations CASCADE;`)
	db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_internal CASCADE;`)
	db.AutoCommit(false)
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

func (d *dummyTxApp) Price(ctx context.Context, dbTx sql.DB, tx *ktypes.Transaction, chainContext *common.ChainContext) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (d *dummyTxApp) Commit() error {
	return nil
}

func (d *dummyTxApp) Rollback() {}

func (d *dummyTxApp) GenesisInit(ctx context.Context, db sql.DB, validators []*ktypes.Validator, genesisAccounts []*ktypes.Account, initialHeight int64, chain *common.ChainContext) error {
	return nil
}

func (d *dummyTxApp) AccountInfo(ctx context.Context, dbTx sql.DB, identifier []byte, pending bool) (*big.Int, int64, error) {
	return big.NewInt(0), 0, nil
}

func (d *dummyTxApp) ApplyMempool(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) error {
	return nil
}

/*type validatorStore struct {
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
}*/

type mockAccounts struct{}

func (m *mockAccounts) Updates() []*ktypes.Account {
	return nil
}

func (m *mockAccounts) GetBlockSpends() []*accounts.Spend {
	return nil
}

type mockEventStore struct {
	events []*ktypes.VotableEvent
}

func (m *mockEventStore) MarkBroadcasted(ctx context.Context, ids []*ktypes.UUID) error {
	return nil
}

func (m *mockEventStore) GetUnbroadcastedEvents(ctx context.Context) ([]*ktypes.UUID, error) {
	var ids []*ktypes.UUID
	for _, event := range m.events {
		ids = append(ids, event.ID())
	}
	return ids, nil
}

// TODO: can test with real migrator
/*type mockMigrator struct{}

func (m *mockMigrator) NotifyHeight(ctx context.Context, block *common.BlockContext, db migrations.Database) error {
	return nil
}

func (m *mockMigrator) StoreChangesets(height int64, changes <-chan any) error {
	return nil
}

func (m *mockMigrator) PersistLastChangesetHeight(ctx context.Context, tx sql.Executor) error {
	return nil
}

func (m *mockMigrator) GetMigrationMetadata(ctx context.Context, status ktypes.MigrationStatus) (*ktypes.MigrationMetadata, error) {
	return nil, nil
}*/
