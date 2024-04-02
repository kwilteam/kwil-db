package abci

import (
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// groupTransactions groups the transactions by sender.
func groupTxsBySender(txns [][]byte) (map[string][]*transactions.Transaction, error) {
	grouped := make(map[string][]*transactions.Transaction)
	for _, tx := range txns {
		t := &transactions.Transaction{}
		err := t.UnmarshalBinary(tx)
		if err != nil {
			return nil, err
		}
		key := string(t.Sender)
		grouped[key] = append(grouped[key], t)
	}
	return grouped, nil
}

// nonceList is for debugging
func nonceList(txns []*transactions.Transaction) []uint64 {
	nonces := make([]uint64, len(txns))
	for i, tx := range txns {
		nonces[i] = tx.Body.Nonce
	}
	return nonces
}
