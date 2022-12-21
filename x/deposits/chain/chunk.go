package chain

import (
	"context"
	"database/sql"
	"kwil/x/deposits/repository"
	"kwil/x/logx"
	"sync"

	"go.uber.org/zap/zapcore"
)

type chunk struct {
	chainClient ChainClient
	dao         *repository.Queries
	tx          *sql.Tx
	lock        *sync.Mutex
	log         logx.SugaredLogger
	start       int64
	finish      int64
}

func (c *chain) newChunk(ctx context.Context, tx *sql.Tx, start, finish int64) (*chunk, error) {
	logger := logx.New().Named("chunk").With(zapcore.Field{
		Key:     "start",
		Type:    zapcore.Int64Type,
		Integer: start,
	}, zapcore.Field{
		Key:     "finish",
		Type:    zapcore.Int64Type,
		Integer: finish,
	}).Sugar()

	return &chunk{
		chainClient: c.chainClient,
		dao:         c.dao.WithTx(tx),
		lock:        &sync.Mutex{},
		start:       start,
		finish:      finish,
		log:         logger,
	}, nil
}
