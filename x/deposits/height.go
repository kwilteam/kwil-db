package deposits

import (
	"context"
	"kwil/kwil/repository"
)

func (s *depositer) GetHeight(ctx context.Context) (int64, error) {
	return s.dao.GetHeight(ctx, int32(s.chain.ChainCode()))
}

func (s *depositer) SetHeight(ctx context.Context, height int64) error {
	return s.dao.SetHeight(ctx, &repository.SetHeightParams{
		Height: height,
		ID:     int32(s.chain.ChainCode()),
	})
}
