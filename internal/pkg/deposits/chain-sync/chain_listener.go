package chainsync

import (
	"context"
	"fmt"
	"kwil/internal/pkg/deposits/tasks"
	escrowtasks "kwil/internal/pkg/deposits/tasks/escrow-tasks"
	"kwil/internal/repository"
	chainClient "kwil/pkg/chain/client"
	"kwil/pkg/chain/contracts/escrow"
	"kwil/pkg/chain/contracts/escrow/types"
	chains "kwil/pkg/chain/types"
	"kwil/pkg/log"
	"kwil/pkg/sql/sqlclient"
	"os"
	"sync"

	"github.com/cstockton/go-conv"
)

type Chain interface {
	RegisterTask(task tasks.Runnable)
	Start(ctx context.Context) error
	ReturnFunds(ctx context.Context, params *types.ReturnFundsParams) (*types.ReturnFundsResponse, error)
	ChainCode() chains.ChainCode
}

type chain struct {
	db             *sqlclient.DB           // for creating new txs
	dao            repository.Queries      // for interacting with the db
	chainClient    chainClient.ChainClient // for getting blocks
	escrowContract escrow.EscrowContract   // for returning funds
	log            log.Logger
	tasks          tasks.TaskRunner
	chunkSize      int64
	mu             *sync.Mutex
	height         int64 // the height of the last block we processed
}

func New(client chainClient.ChainClient, escrow escrow.EscrowContract, dao repository.Queries, db *sqlclient.DB, logger log.Logger, providerAddress string) (Chain, error) {
	escrowTasks := escrowtasks.New(dao, escrow, providerAddress)
	chunkSizeEnv := os.Getenv("deposit_chunk_size")
	if chunkSizeEnv == "" {
		chunkSizeEnv = "100000"
	}
	chunkSize, err := conv.Int64(chunkSizeEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chunk size to int64: %w", err)
	}

	// create the task runner with escrow tasks
	commitHeightTask := tasks.NewHeightTask(dao, client.ChainCode())
	taskRunner := tasks.New(escrowTasks)

	// set the final task to be commit height
	taskRunner.SetFinal(commitHeightTask)

	return &chain{
		db:             db,
		dao:            dao,
		chainClient:    client,
		escrowContract: escrow,
		log:            logger.Named("deposit.chain"),
		chunkSize:      chunkSize,
		mu:             &sync.Mutex{},
		height:         0,
		tasks:          taskRunner,
	}, nil
}
