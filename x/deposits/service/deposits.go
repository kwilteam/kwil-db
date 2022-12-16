package service

import (
	"context"
	"kwil/x/deposits/dto"
)

type DepositsService interface {
	Spend(ctx context.Context, spend dto.Spend) error
}

type depositsService struct {
}

func NewService() DepositsService {
	return &depositsService{}
}

func (s *depositsService) Spend(ctx context.Context, spend dto.Spend) error {
	panic("implement me")
}
