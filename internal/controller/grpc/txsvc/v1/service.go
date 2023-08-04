package txsvc

import (
	"context"
	"math/big"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/balances"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	engine       EngineReader
	accountStore AccountReader

	cometBftClient BlockchainBroadcaster
}

func NewService(engine EngineReader, accountStore AccountReader, cometBftClient BlockchainBroadcaster, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:            log.NewNoOp(),
		engine:         engine,
		accountStore:   accountStore,
		cometBftClient: cometBftClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type EngineReader interface {
	Call(ctx context.Context, call *tx.CallActionPayload, msg *tx.SignedMessage[tx.JsonPayload]) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
	ListOwnedDatabases(ctx context.Context, owner string) ([]string, error)
	PriceDeploy(ctx context.Context, schema *engineTypes.Schema) (price *big.Int, err error)
	PriceDrop(ctx context.Context, dbid string) (price *big.Int, err error)
	PriceExecute(ctx context.Context, dbid string, action string, params []map[string]any) (price *big.Int, err error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
}

type AccountReader interface {
	GetAccount(ctx context.Context, address string) (*balances.Account, error)
}

type BlockchainBroadcaster interface {
	// TODO: this should be refactored to: BroadcastTxAsync(ctx context.Context, tx tx.Transaction) error
	// this will remove abci and cometbft as a dependency from this package, and functionally works the same
	BroadcastTxAsync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error)
}
