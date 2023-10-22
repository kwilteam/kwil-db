package txsvc

import (
	"context"
	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/kwilteam/kwil-db/core/log"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/validators"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger
	// chainID is the Kwil network ID. This is static, so we'll set it in the
	// constructor even though there are ways to get it via the node app or chain client.
	chainID string

	engine       EngineReader
	accountStore AccountReader
	vstore       ValidatorReader

	chainClient BlockchainTransactor
}

func NewService(chainID string, engine EngineReader, accountStore AccountReader, vstore ValidatorReader,
	chainClient BlockchainTransactor, nodeApp NodeApplication, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:          log.NewNoOp(),
		chainID:      chainID,
		engine:       engine,
		accountStore: accountStore,
		vstore:       vstore,
		chainClient:  chainClient,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type EngineReader interface {
	Call(ctx context.Context, dbid string, action string, args []any, msg *transactions.CallMessage) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
	ListOwnedDatabases(ctx context.Context, owner []byte) ([]string, error)
	PriceDeploy(ctx context.Context, schema *engineTypes.Schema) (price *big.Int, err error)
	PriceDrop(ctx context.Context, dbid string) (price *big.Int, err error)
	PriceExecute(ctx context.Context, dbid string, action string, args [][]any) (price *big.Int, err error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
}

type AccountReader interface {
	GetAccount(ctx context.Context, pubkey []byte) (*accounts.Account, error)
}

type BlockchainTransactor interface {
	Status(ctx context.Context) (*types.Status, error)
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (txHash []byte, err error)
	TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error) // TODO: don't use comet types here
}

type NodeApplication interface {
	ChainID() string
}

type ValidatorReader interface {
	CurrentValidators(ctx context.Context) ([]*validators.Validator, error)
	ActiveVotes(ctx context.Context) ([]*validators.JoinRequest, error)
	// JoinStatus(ctx context.Context, joiner []byte) ([]*JoinRequest, error)
	PriceJoin(ctx context.Context) (*big.Int, error)
	PriceLeave(ctx context.Context) (*big.Int, error)
	PriceApprove(ctx context.Context) (*big.Int, error)
}
