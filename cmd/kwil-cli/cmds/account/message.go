package account

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

type respAccount struct {
	Identifier types.HexBytes `json:"identifier"`
	KeyType    string         `json:"key_type"`
	Balance    string         `json:"balance"`
	Nonce      int64          `json:"nonce"`
}

func (r *respAccount) MarshalJSON() ([]byte, error) {
	type respAccountAlias respAccount
	return json.Marshal((*respAccountAlias)(r))
}

func (r *respAccount) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf(`%x (%s)
Balance: %s
Nonce: %d
`, r.Identifier, r.KeyType, r.Balance, r.Nonce)

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
