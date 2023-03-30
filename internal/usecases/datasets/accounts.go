package datasets

import "kwil/internal/entity"

func (u *DatasetUseCase) GetAccount(address string) (*entity.Account, error) {
	acc, err := u.accountStore.GetAccount(address)
	if err != nil {
		return nil, err
	}

	return &entity.Account{
		Address: acc.Address,
		Balance: acc.Balance.String(),
		Nonce:   acc.Nonce,
	}, nil
}
