package txsvc

import (
	"context"
	"math/big"
	"time"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
	coreTypes "github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

const defaultReadTxTimeout = 5 * time.Second

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	readTxTimeout time.Duration
	engine        EngineReader

	db sql.ReadTxMaker // this should only ever make a read-only tx

	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	pricer      Pricer
	chainClient BlockchainTransactor
}

func NewService(db sql.ReadTxMaker, engine EngineReader,
	chainClient BlockchainTransactor, nodeApp NodeApplication, pricer Pricer, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:           log.NewNoOp(),
		readTxTimeout: defaultReadTxTimeout,
		engine:        engine,
		nodeApp:       nodeApp,
		pricer:        pricer,
		chainClient:   chainClient,
		db:            db,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type EngineReader interface {
	Procedure(ctx context.Context, tx sql.DB, options *common.ExecutionData) (*sql.ResultSet, error)
	GetSchema(dbid string) (*types.Schema, error)
	ListDatasets(owner []byte) ([]*coreTypes.DatasetIdentifier, error)
	Execute(ctx context.Context, tx sql.DB, dbid string, query string, values map[string]any) (*sql.ResultSet, error)
}

type BlockchainTransactor interface {
	Status(ctx context.Context) (*adminTypes.Status, error)
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (*cmtCoreTypes.ResultBroadcastTx, error)
	TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error)
}

type NodeApplication interface {
	AccountInfo(ctx context.Context, db sql.DB, identifier []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
}

type Pricer interface {
	Price(ctx context.Context, db sql.DB, tx *transactions.Transaction) (*big.Int, error)
}
