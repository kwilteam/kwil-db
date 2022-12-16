package service

import "context"

func (s *depositsService) GetHeight(ctx context.Context) (int64, error) {
	return s.doa.GetHeight(ctx)
}

func (s *depositsService) SetHeight(ctx context.Context, height int64) error {
	return s.doa.SetHeight(ctx, height)
}
