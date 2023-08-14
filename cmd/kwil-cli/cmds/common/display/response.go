package display

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func PrintTxResponse(res *transactions.TransactionStatus) {
	if res.ID != nil {
		fmt.Println("Success!")
	}
	fmt.Println("Response:")
	fmt.Println("  Hash:", hexutil.Encode(res.ID))
	fmt.Println("  Fee:", res.Fee)
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
