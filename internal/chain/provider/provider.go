package provider

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/chain/provider/evm"
	"github.com/kwilteam/kwil-db/internal/chain/types"
)

func New(endpoint string, chainCode types.ChainCode, tokenAddress string, escrowAddress string) (ChainProvider, error) {
	switch chainCode {
	case types.ETHEREUM, types.GOERLI, types.LOCAL:
		return evm.New(endpoint, chainCode, tokenAddress, escrowAddress)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainCode))
	}
}
