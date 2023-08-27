package display

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func PrintTxResponse(res transactions.TxHash) {
	fmt.Println("Response:")
	fmt.Println("  Hash:", res.Hex())
}

type ClientChainResponse struct {
	Chain string `json:"chain"`
	Tx    string `json:"tx"`
}

func PrintClientChainResponse(res *ClientChainResponse) {
	fmt.Println("Response:")
	fmt.Println("  Chain:", res.Chain)
	fmt.Println("  Tx:", res.Tx)
}
