package datasets

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
)

// compareAndSpend compares the price to the fee and spends the price if the fee is enough.
func (u *DatasetUseCase) compareAndSpend(address, fee string, nonce int64, price *big.Int) error {
	// convert fee to big.Int
	bigFee, ok := big.NewInt(0).SetString(fee, 10)
	if !ok {
		return fmt.Errorf("failed to convert fee to big.Int")
	}

	// compare price to fee
	if price.Cmp(bigFee) > 0 {
		fmt.Println("fee is not enough", price, bigFee)
		return fmt.Errorf("fee is not enough")
	}

	return u.accountStore.Spend(&balances.Spend{
		AccountAddress: address,
		Amount:         price,
		Nonce:          nonce,
	})
}
