package tasks

import (
	"context"
	"kwil/x/chain"
	"kwil/x/deposits/repository"
)

type heightTask struct {
	chainCode chain.ChainCode
	dao       *repository.Queries
	queries   *repository.Queries
}

func NewHeightTask(dao *repository.Queries, chainCode chain.ChainCode) Runnable {
	return &heightTask{
		chainCode: chainCode,
		dao:       nil,
		queries:   dao,
	}
}

func (t *heightTask) Run(ctx context.Context, chunk *Chunk) error {
	t.dao = t.queries.WithTx(chunk.Tx)

	err := t.dao.SetHeight(ctx, &repository.SetHeightParams{
		Height: chunk.Finish,
		ID:     int32(t.chainCode),
	})
	if err != nil {
		return err
	}

	t.dao = nil

	return nil
}
