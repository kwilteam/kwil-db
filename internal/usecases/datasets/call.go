package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
)

func (u *DatasetUseCase) Call(ctx context.Context, action *entity.CallAction) ([]map[string]any, error) {
	ds, err := u.engine.GetDataset(ctx, action.Payload.DBID)
	if err != nil {
		return nil, err
	}

	return ds.Call(ctx, action.Payload.Action, action.Payload.Params, &dataset.TxOpts{
		Caller: action.Message.Sender,
	})
}
