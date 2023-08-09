package datasets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
)

func (u *DatasetUseCase) spend(ctx context.Context, address string, amount *big.Int, nonce int64) error {
	return u.accountStore.Spend(ctx, &balances.Spend{
		AccountAddress: address,
		Amount:         amount,
		Nonce:          nonce,
	})
}

func (u *DatasetUseCase) Spend(ctx context.Context, address string, amount string, nonce int64) error {
	amountBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("failed to parse amount")
	}

	return u.spend(ctx, address, amountBigInt, nonce)
}
