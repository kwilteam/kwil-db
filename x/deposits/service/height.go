package service

import (
	"context"
	"kwil/x/deposits/repository"
)

func (s *depositsService) GetHeight(ctx context.Context) (int64, error) {
	return s.dao.GetHeight(ctx, int32(s.chain.ChainCode()))
}

func (s *depositsService) SetHeight(ctx context.Context, height int64) error {
	return s.dao.SetHeight(ctx, &repository.SetHeightParams{
		Height: height,
		ID:     int32(s.chain.ChainCode()),
	})
}
