package client

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/chain/contracts"
	provider "github.com/kwilteam/kwil-db/pkg/chain/provider/dto"
	"github.com/kwilteam/kwil-db/pkg/chain/types"
)

type ChainClient interface {
	Listen(ctx context.Context, blocks chan<- int64) error
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
	ChainCode() types.ChainCode
	Close() error
	Contracts() contracts.Contracter
}
