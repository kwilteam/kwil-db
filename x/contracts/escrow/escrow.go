package escrow

import (
	"fmt"
	"kwil/x/chain"
	provider "kwil/x/chain/provider/dto"
	"kwil/x/contracts/escrow/dto"
	"kwil/x/contracts/escrow/evm"
)

func New(provider provider.ChainProvider, privateKey, address string) (dto.EscrowContract, error) {
	switch provider.ChainCode() {
	case chain.ETHEREUM:
		return evm.New(provider, privateKey, address)
	case chain.GOERLI:
		return evm.New(provider, privateKey, address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(provider.ChainCode()))
	}
}
