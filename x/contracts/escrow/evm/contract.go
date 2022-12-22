package evm

import (
	"crypto/ecdsa"
	"kwil/abi"
	"kwil/x/contracts/escrow/dto"
	"kwil/x/crypto"
	"math/big"

	providerDTO "kwil/x/chain/provider/dto"

	"github.com/ethereum/go-ethereum/common"
)

type contract struct {
	ctr         *abi.Escrow
	token       string
	cid         *big.Int
	key         *ecdsa.PrivateKey
	nodeAddress string
}

func New(provider providerDTO.ChainProvider, privateKey, contractAddress string) (dto.EscrowContract, error) {
	client, err := provider.AsEthClient()
	if err != nil {
		return nil, err
	}

	ctr, err := abi.NewEscrow(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, err
	}

	tokAddr, err := ctr.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	// private key to address
	nodeAddress, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	// get hex private key
	pKeyHex, err := crypto.ECDSAFromHex(privateKey)
	if err != nil {
		return nil, err
	}

	return &contract{
		ctr:         ctr,
		token:       tokAddr.Hex(),
		cid:         provider.ChainCode().ToChainId(),
		key:         pKeyHex,
		nodeAddress: nodeAddress,
	}, nil
}
