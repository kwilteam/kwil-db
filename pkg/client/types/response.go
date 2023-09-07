package types

import "github.com/kwilteam/kwil-db/pkg/transactions"

type TxQueryResponse struct {
	Hash     []byte                          `json:"hash"`
	Height   int64                           `json:"height"`
	Tx       *transactions.Transaction       `json:"tx"`
	TxResult *transactions.TransactionResult `json:"tx_result"`
}
