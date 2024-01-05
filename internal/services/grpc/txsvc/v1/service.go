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
	"github.com/kwilteam/kwil-db/internal/accounts"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/validators"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger
	// chainID is the Kwil network ID.
	chainID string

	engine       EngineReader
	accountStore AccountReader
	vstore       ValidatorReader

	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	chainClient BlockchainTransactor
}

func NewService(engine EngineReader, accountStore AccountReader, vstore ValidatorReader,
	chainClient BlockchainTransactor, nodeApp NodeApplication, opts ...TxSvcOpt) *Service {
	s := &Service{
		log:          log.NewNoOp(),
		chainID:      nodeApp.ChainID(),
		engine:       engine,
		accountStore: accountStore,
		vstore:       vstore,
		nodeApp:      nodeApp,
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
	ListOwnedDatabases(ctx context.Context, owner []byte) ([]*coreTypes.DatasetIdentifier, error)
	PriceDeploy(ctx context.Context, schema *engineTypes.Schema) (price *big.Int, err error)
	PriceDrop(ctx context.Context, dbid string) (price *big.Int, err error)
	PriceExecute(ctx context.Context, dbid string, action string, args [][]any) (price *big.Int, err error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
}

type AccountReader interface {
	Account(ctx context.Context, identifier []byte) (*accounts.Account, error)
	PriceTransfer(ctx context.Context) (*big.Int, error)
}

type BlockchainTransactor interface {
	Status(ctx context.Context) (*adminTypes.Status, error)
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (*cmtCoreTypes.ResultBroadcastTx, error)
	TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error)
}

type NodeApplication interface {
	ChainID() string
	AccountInfo(ctx context.Context, identifier []byte) (balance *big.Int, nonce int64, err error)
}

type ValidatorReader interface {
	CurrentValidators(ctx context.Context) ([]*validators.Validator, error)
	ActiveVotes(ctx context.Context) ([]*validators.JoinRequest, []*validators.ValidatorRemoveProposal, error)
	// JoinStatus(ctx context.Context, joiner []byte) ([]*JoinRequest, error)
	PriceJoin(ctx context.Context) (*big.Int, error)
	PriceLeave(ctx context.Context) (*big.Int, error)
	PriceApprove(ctx context.Context) (*big.Int, error)
	PriceRemove(ctx context.Context) (*big.Int, error)
}
