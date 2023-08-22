package txsvc

import (
	"context"
	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/balances"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	engine       EngineReader
	accountStore AccountReader

	chainClient BlockchainBroadcaster
}

func NewService(engine EngineReader, accountStore AccountReader, chainClient BlockchainBroadcaster, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:          log.NewNoOp(),
		engine:       engine,
		accountStore: accountStore,
		chainClient:  chainClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type EngineReader interface {
	Call(ctx context.Context, dbid string, action string, args []any, msg *transactions.SignedMessage) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
	ListOwnedDatabases(ctx context.Context, owner string) ([]string, error)
	PriceDeploy(ctx context.Context, schema *engineTypes.Schema) (price *big.Int, err error)
	PriceDrop(ctx context.Context, dbid string) (price *big.Int, err error)
	PriceExecute(ctx context.Context, dbid string, action string, args [][]any) (price *big.Int, err error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
}

type AccountReader interface {
	GetAccount(ctx context.Context, address string) (*balances.Account, error)
}

type BlockchainBroadcaster interface {
	BroadcastTxAsync(ctx context.Context, tx *transactions.Transaction) (txHash []byte, err error)
	TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error)
}
