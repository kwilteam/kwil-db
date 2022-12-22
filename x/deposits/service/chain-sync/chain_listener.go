package chainsync

import (
	"context"
	chainClientDTO "kwil/x/chain/client/dto"
	escrowDTO "kwil/x/contracts/escrow/dto"
	"kwil/x/deposits/repository"
	"kwil/x/logx"
	"kwil/x/sqlx/sqlclient"
	"sync"

	"kwil/x/deposits/service/tasks"
	escrowtasks "kwil/x/deposits/service/tasks/escrow-tasks"
)

type Chain interface {
	RegisterTask(task tasks.Runnable)
	Start(ctx context.Context) error
	ReturnFunds(ctx context.Context, params *escrowDTO.ReturnFundsParams) (*escrowDTO.ReturnFundsResponse, error)
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
	height         int64
}

func New(client chainClientDTO.ChainClient, escrow escrowDTO.EscrowContract, dao *repository.Queries, db *sqlclient.DB) Chain {
	escrowTasks := escrowtasks.New(dao, escrow)
	taskRunner := tasks.New(escrowTasks)
	return &chain{
		db:             db,
		dao:            repository.New(db),
		chainClient:    client,
		escrowContract: escrow,
		log:            logx.New().Named("deposit-chain-client").Sugar(),
		chunkSize:      100000,
		mu:             &sync.Mutex{},
		height:         0,
		tasks:          taskRunner,
	}
}
