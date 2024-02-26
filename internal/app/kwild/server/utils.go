package server

import (
	"context"
	"fmt"
	"os"

	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft/privval"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sessions"
	sqlSessions "github.com/kwilteam/kwil-db/pkg/sessions/sql-session"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/client"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

var defaultFilePath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultFilePath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

func init() {}

// connectExtensions connects to the provided extension urls.
func connectExtensions(ctx context.Context, urls []string) (map[string]*extensions.Extension, error) {
	exts := make(map[string]*extensions.Extension, len(urls))

	for _, url := range urls {
		ext := extensions.New(url)
		err := ext.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect extension '%s': %w", ext.Name(), err)
		}

		_, ok := exts[ext.Name()]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", ext.Name())
		}

		exts[ext.Name()] = ext
	}

	return exts, nil
}

func adaptExtensions(exts map[string]*extensions.Extension) map[string]engine.ExtensionInitializer {
	adapted := make(map[string]engine.ExtensionInitializer, len(exts))

	for name, ext := range exts {
		adapted[name] = extensionInitializeFunc(ext.CreateInstance)
	}

	return adapted
}

type extensionInitializeFunc func(ctx context.Context, metadata map[string]string) (*extensions.Instance, error)

func (e extensionInitializeFunc) CreateInstance(ctx context.Context, metadata map[string]string) (engine.ExtensionInstance, error) {
	return e(ctx, metadata)
}

type sqliteOpener struct {
	sqliteFilePath string
}

func newSqliteOpener(sqliteFilePath string) *sqliteOpener {
	if sqliteFilePath == "" {
		sqliteFilePath = defaultFilePath
	}

	return &sqliteOpener{
		sqliteFilePath: sqliteFilePath,
	}
}

func (s *sqliteOpener) Open(fileName string, logger log.Logger) (sql.Database, error) {
	return client.NewSqliteStore(fileName,
		client.WithLogger(logger),
		client.WithPath(s.sqliteFilePath),
	)
}

// wrappedCometBFTClient satisfies the generic txsvc.BlockchainBroadcaster
// interface, hiding the details of cometBFT.
type wrappedCometBFTClient struct {
	*cmtlocal.Local
}

func (wc *wrappedCometBFTClient) BroadcastTxAsync(ctx context.Context, tx *transactions.Transaction) error {
	bts, err := tx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to serialize transaction data: %w", err)
	}

	_, err = wc.Local.BroadcastTxAsync(ctx, cmttypes.Tx(bts))
	if err != nil {
		return err
	}

	return err
}

// atomicReadWriter implements the CometBFt AtomicReadWriter interface.
// This should probably be done with a file instead of a KV store,
// but we already have a good implementation of an atomic KV store.
type atomicReadWriter struct {
	kv  kv.KVStore
	key []byte
}

var _ privval.AtomicReadWriter = (*atomicReadWriter)(nil)

func (a *atomicReadWriter) Read() ([]byte, error) {
	res, err := a.kv.Get(a.key)
	if err == kv.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (a *atomicReadWriter) Write(val []byte) error {
	return a.kv.Set(a.key, val)
}

// sqlCommittableRegister allows dynamic registration of SQL committables
// it implements engine.CommitRegister
type sqlCommittableRegister struct {
	committer *sessions.AtomicCommitter
	log       log.Logger
}

var _ engine.CommitRegister = (*sqlCommittableRegister)(nil)

func (s *sqlCommittableRegister) Register(ctx context.Context, name string, db sql.Database) error {
	return registerSQL(ctx, s.committer, db, name, s.log)
}

func (s *sqlCommittableRegister) Unregister(ctx context.Context, name string) error {
	return s.committer.Unregister(ctx, name)
}

// registerSQL is a helper function to register a SQL committable to the atomic committer.
func registerSQL(ctx context.Context, ac *sessions.AtomicCommitter, db sql.Database, name string, logger log.Logger) error {
	return ac.Register(ctx, name,
		sqlSessions.NewSqlCommitable(db,
			sqlSessions.WithLogger(*logger.Named(name + "-committable")),
		),
	)
}
