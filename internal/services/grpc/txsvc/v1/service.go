package txsvc

import (
	"context"
	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/kwilteam/kwil-db/core/log"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	coreTypes "github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	engine EngineReader

	db sql.OuterTxMaker // this should only ever make a read-only tx

	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	chainClient BlockchainTransactor
}

func NewService(db sql.OuterTxMaker, engine EngineReader,
	chainClient BlockchainTransactor, nodeApp NodeApplication, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:         log.NewNoOp(),
		engine:      engine,
		nodeApp:     nodeApp,
		chainClient: chainClient,
		db:          db,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type EngineReader interface {
	Execute(ctx context.Context, tx sql.DB, options *engineTypes.ExecutionData) (*sql.ResultSet, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
	ListDatasets(ctx context.Context, owner []byte) ([]*coreTypes.DatasetIdentifier, error)
	Query(ctx context.Context, tx sql.DB, dbid string, query string) (*sql.ResultSet, error)
}

type BlockchainTransactor interface {
	Status(ctx context.Context) (*adminTypes.Status, error)
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (*cmtCoreTypes.ResultBroadcastTx, error)
	TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error)
}

type NodeApplication interface {
	AccountInfo(ctx context.Context, identifier []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
	Price(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
}

// readTx creates a new read-only transaction for the database.
func (s *Service) readTx(ctx context.Context) (sql.Tx, error) {
	outer, err := s.db.BeginTx(ctx, sql.ReadOnly)
	if err != nil {
		return nil, err
	}

	return &wrappedTx{tx: outer}, nil
}

// wrappedTx wraps an OuterTx to abstract away the precommit method.
// This is necessary because the outermost tx needs to do the precommit,
// but the engine is not aware of precommit.
type wrappedTx struct {
	tx sql.OuterTx
}

func (w *wrappedTx) AccessMode() sql.AccessMode {
	return w.tx.AccessMode()
}

func (w *wrappedTx) BeginSavepoint(ctx context.Context) (sql.Tx, error) {
	return w.tx.BeginSavepoint(ctx)
}

func (w *wrappedTx) Commit(ctx context.Context) error {
	_, err := w.tx.Precommit(ctx)
	if err != nil {
		return err
	}

	return w.tx.Commit(ctx)
}

func (w *wrappedTx) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return w.tx.Execute(ctx, stmt, args...)
}

func (w *wrappedTx) Rollback(ctx context.Context) error {
	return w.tx.Rollback(ctx)
}
