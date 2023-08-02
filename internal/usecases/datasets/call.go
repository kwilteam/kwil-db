package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
)

func (u *DatasetUseCase) Call(ctx context.Context, action *entity.ActionCall) ([]map[string]any, error) {
	ds, err := u.engine.GetDataset(ctx, action.Message.Payload.DBID)
	if err != nil {
		return nil, err
	}

	return ds.Call(ctx, action.Message.Payload.Action, action.Message.Payload.Params, &dataset.TxOpts{
		Caller: action.Message.Sender,
	})
}
