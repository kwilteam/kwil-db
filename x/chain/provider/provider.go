package provider

import (
	"fmt"
	"kwil/x/chain"
	"kwil/x/chain/provider/dto"
	"kwil/x/chain/provider/evm"
)

func New(endpoint string, chainCode chain.ChainCode) (dto.ChainProvider, error) {
	switch chainCode {
	case chain.ETHEREUM:
		return evm.New(endpoint, chainCode)
	case chain.GOERLI:
		return evm.New(endpoint, chainCode)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainCode))
	}
}
