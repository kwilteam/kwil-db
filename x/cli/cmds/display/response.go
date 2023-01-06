package display

import (
	"fmt"
	txDto "kwil/x/transactions/dto"
)

func PrintTxResponse(res *txDto.Response) {
	fmt.Println("Response:")
	fmt.Println("  Hash:", res.Hash)
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
