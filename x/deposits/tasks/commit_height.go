package tasks

import (
	"context"
	"kwil/kwil/repository"
	"kwil/pkg/chain/types"
)

type heightTask struct {
	chainCode types.ChainCode
	dao       repository.Queries
	queries   repository.Queries
}

func NewHeightTask(dao repository.Queries, chainCode types.ChainCode) Runnable {
	return &heightTask{
		chainCode: chainCode,
		dao:       nil,
		queries:   dao,
	}
}

func (t *heightTask) Run(ctx context.Context, chunk *Chunk) error {
	t.dao = t.queries.WithTx(chunk.Tx)

	err := t.dao.SetHeight(ctx, int32(t.chainCode), chunk.Finish)
	if err != nil {
		return err
	}

	t.dao = nil

	return nil
}
