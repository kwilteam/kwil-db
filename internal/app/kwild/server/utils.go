package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/abci"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft/privval"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sessions"
	sqlSessions "github.com/kwilteam/kwil-db/pkg/sessions/sql-session"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/client"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"
	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

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

func (wc *wrappedCometBFTClient) BroadcastTx(ctx context.Context, tx []byte, sync uint8) (uint32, []byte, error) {
	var bcastFun func(ctx context.Context, tx cmttypes.Tx) (*cmtCoreTypes.ResultBroadcastTx, error)
	switch sync {
	case 0:
		bcastFun = wc.Local.BroadcastTxAsync
	case 1:
		bcastFun = wc.Local.BroadcastTxSync
	case 2:
		bcastFun = func(ctx context.Context, tx cmttypes.Tx) (*cmtCoreTypes.ResultBroadcastTx, error) {
			res, err := wc.Local.BroadcastTxCommit(ctx, tx)
			if err != nil {
				return nil, err
			}
			if res.CheckTx.Code != abciTypes.CodeTypeOK {
				return &cmtCoreTypes.ResultBroadcastTx{
					Code:      res.CheckTx.Code,
					Data:      res.CheckTx.Data,
					Log:       res.CheckTx.Log,
					Codespace: res.CheckTx.Codespace,
					Hash:      res.Hash,
				}, nil
			}
			return &cmtCoreTypes.ResultBroadcastTx{
				Code:      res.DeliverTx.Code,
				Data:      res.DeliverTx.Data,
				Log:       res.DeliverTx.Log,
				Codespace: res.DeliverTx.Codespace,
				Hash:      res.Hash,
			}, nil
		}
	}

	result, err := bcastFun(ctx, cmttypes.Tx(tx))
	if err != nil {
		return 0, nil, err
	}

	return result.Code, result.Hash.Bytes(), nil
}

// TxQuery locates a transaction in the node's blockchain or mempool. If the
// transaction could not be located, and error of type pkg/abci.ErrTxNotFound is
// returned.
func (wc *wrappedCometBFTClient) TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error) {
	// First check confirmed transactions. The Tx method of the cometbft client
	// does not define a specific exported error for a transaction that is not
	// found, just a "tx (%X) not found" as of cometbft v0.37. The Tx docs also
	// indicate that "`nil` could mean the transaction is in the mempool", so
	// this API should be used with caution, not failing on error AND checking
	// the result for nilness.
	res, err := wc.Local.Tx(ctx, hash, prove)
	if err == nil && res != nil {
		return res, nil
	}
	// The transaction could be in mempool.
	limit := -1
	unconf, err := wc.Local.UnconfirmedTxs(ctx, &limit)
	if err != nil {
		return nil, err
	}
	for _, tx := range unconf.Txs {
		if bytes.Equal(tx.Hash(), hash) {
			// Found it. Shoe-horn into a ResultTx with -1 height, and the zero
			// values for ResponseDeliverTx and TxProof (it's checked and
			// accepted to mempool, but not delivered in a block yet).
			return &cmtCoreTypes.ResultTx{
				Hash:   hash,
				Height: -1,
				Tx:     tx,
			}, nil
		}
	}
	return nil, abci.ErrTxNotFound
}

// atomicReadWriter implements the CometBFT AtomicReadWriter interface.
// This should probably be done with a file instead of a KV store,
// but we already have a good implementation of an atomic KV store.
type atomicReadWriter struct {
	kv  kv.KVStore
	key []byte
}

var _ privval.AtomicReadWriter = (*atomicReadWriter)(nil)

func (a *atomicReadWriter) Read() ([]byte, error) {
	res, err := a.kv.Get(a.key)
	if errors.Is(err, kv.ErrKeyNotFound) {
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
