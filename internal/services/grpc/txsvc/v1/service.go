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
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	engine EngineReader

	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	chainClient BlockchainTransactor
}

func NewService(engine EngineReader,
	chainClient BlockchainTransactor, nodeApp NodeApplication, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:         log.NewNoOp(),
		engine:      engine,
		nodeApp:     nodeApp,
		chainClient: chainClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type EngineReader interface {
	Call(ctx context.Context, dbid string, action string, args []any, msg *transactions.CallMessage) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
	ListDatasets(ctx context.Context, owner []byte) ([]*coreTypes.DatasetIdentifier, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
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
