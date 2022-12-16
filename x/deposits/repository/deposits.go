package repository

import (
	"context"
	"kwil/x/deposits/dto"
)

type DepositQuery interface {
}

type depositQuery struct{}

func (d *depositQuery) Spend(ctx context.Context, spend dto.Spend) error {
	panic("implement me")
}
