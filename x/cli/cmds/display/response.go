package display

import (
	"fmt"
	txDto "kwil/x/transactions/dto"
)

func PrintResponse(res *txDto.Response) {
	fmt.Println("Response:")
	fmt.Println("  Hash:", res.Hash)
	fmt.Println("  Fee:", res.Fee)
}
