package account

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

type respAccount types.Account

func (r *respAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Identifier string `json:"identifier"`
		Balance    string `json:"balance"`
		Nonce      int64  `json:"nonce"`
	}{
		Identifier: hex.EncodeToString(r.Identifier),
		Balance:    r.Balance.String(),
		Nonce:      r.Nonce,
	})
}

func (r *respAccount) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf(`Account ID: %x
Balance: %s
Nonce: %d
`, r.Identifier, r.Balance, r.Nonce)

	return []byte(msg), nil
}

/*xxx
type respAccount struct {
	// Identifier string `json:"identifier"`
	Balance string `json:"balance"`
	Nonce   int64  `json:"nonce"`
}

func (r *respAccount) MarshalJSON() ([]byte, error) {
	type respAccountAlias respAccount // avoid infinite recursion
	return json.Marshal((*respAccountAlias)(r))
}

func (r *respAccount) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf(`Balance: %s
Nonce: %d
`, r.Balance, r.Nonce)

	return []byte(msg), nil
}
*/
