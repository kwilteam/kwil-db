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

	db sql.ReadTxMaker // this should only ever make a read-only tx

	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	chainClient BlockchainTransactor
}

func NewService(db sql.ReadTxMaker, engine EngineReader,
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
