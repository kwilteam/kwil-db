package display

import (
	"fmt"
	txTypes "github.com/kwilteam/kwil-db/pkg/tx"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func PrintTxResponse(res *txTypes.Receipt) {
	if res.TxHash != nil {
		fmt.Println("Success!")
	}
	fmt.Println("Response:")
	fmt.Println("  Hash:", hexutil.Encode(res.TxHash))
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
