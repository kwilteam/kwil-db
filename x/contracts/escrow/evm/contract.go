package evm

import (
	"crypto/ecdsa"
	"kwil/abi"
	"kwil/x/contracts/escrow/dto"
	"kwil/x/crypto"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	ctr         *abi.Escrow
	token       string
	cid         *big.Int
	key         *ecdsa.PrivateKey
	nodeAddress string
}

func New(client *ethclient.Client, chainId *big.Int, privateKey, contractAddress string) (dto.EscrowContract, error) {

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
		cid:         chainId,
		key:         pKeyHex,
		nodeAddress: nodeAddress,
	}, nil
}
