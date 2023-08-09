package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/entity"
)

func (u *DatasetUseCase) GetAccount(ctx context.Context, address string) (*entity.Account, error) {
	acc, err := u.accountStore.GetAccount(ctx, address)
	if err != nil {
		return nil, err
	}

	return &entity.Account{
		Address: acc.Address,
		Balance: acc.Balance.String(),
		Nonce:   acc.Nonce,
	}, nil
}
