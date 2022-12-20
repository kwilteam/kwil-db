package builder

import (
	evmclient "kwil/x/chain-client/evm/client"
	evmescrow "kwil/x/chain-client/evm/contracts/escrow"
	"math/big"

	ethc "github.com/ethereum/go-ethereum/ethclient"
)

func (b *chainBuilder) buildEVM(chainId *big.Int) (*ChainComponents, error) {
	client, err := ethc.Dial(b.rpc)
	if err != nil {
		return nil, err
	}

	escrowContract, err := evmescrow.New(client, chainId, b.privateKey, b.contractAddress)
	if err != nil {
		return nil, err
	}

	return &ChainComponents{
		Escrow:   escrowContract,
		Listener: evmclient.New(client, chainId),
	}, nil
}
