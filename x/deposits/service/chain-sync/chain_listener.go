package chainsync

import (
	"context"
	"fmt"
	chains "kwil/x/chain"
	chainClientDTO "kwil/x/chain/client/dto"
	escrowDTO "kwil/x/contracts/escrow/dto"
	"kwil/x/deposits/repository"
	"kwil/x/logx"
	"kwil/x/sqlx/sqlclient"
	"os"
	"sync"

	"kwil/x/deposits/service/tasks"
	escrowtasks "kwil/x/deposits/service/tasks/escrow-tasks"

	"github.com/cstockton/go-conv"
)

type Chain interface {
	RegisterTask(task tasks.Runnable)
	Start(ctx context.Context) error
	ReturnFunds(ctx context.Context, params *escrowDTO.ReturnFundsParams) (*escrowDTO.ReturnFundsResponse, error)
	ChainCode() chains.ChainCode
}

type chain struct {
	db             *sqlclient.DB              // for creating new txs
	dao            *repository.Queries        // for interacting with the db
	chainClient    chainClientDTO.ChainClient // for getting blocks
	escrowContract escrowDTO.EscrowContract   // for returning funds
	log            logx.SugaredLogger
	tasks          tasks.TaskRunner
	chunkSize      int64
	mu             *sync.Mutex
	height         int64 // the height of the last block we processed
}

func New(client chainClientDTO.ChainClient, escrow escrowDTO.EscrowContract, dao *repository.Queries, db *sqlclient.DB) (Chain, error) {
	escrowTasks := escrowtasks.New(dao, escrow)
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
		dao:            repository.New(db),
		chainClient:    client,
		escrowContract: escrow,
		log:            logx.New().Named("deposit-chain-client").Sugar(),
		chunkSize:      chunkSize,
		mu:             &sync.Mutex{},
		height:         0,
		tasks:          taskRunner,
	}, nil
}
