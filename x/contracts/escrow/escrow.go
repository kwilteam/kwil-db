package escrow

import (
	"fmt"
	"kwil/x/chain"
	ccDTO "kwil/x/chain/client/dto"
	"kwil/x/contracts/escrow/dto"
	"kwil/x/contracts/escrow/evm"
)

func New(chainClient ccDTO.ChainClient, privateKey, address string) (dto.EscrowContract, error) {
	switch chainClient.ChainCode() {
	case chain.ETHEREUM, chain.GOERLI:
		ethClient, err := chainClient.AsEthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get ethclient from chain client: %d", err)
		}

		return evm.New(ethClient, chainClient.ChainCode().ToChainId(), privateKey, address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainClient.ChainCode()))
	}
}
