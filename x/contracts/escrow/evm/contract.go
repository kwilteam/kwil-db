package evm

import (
	"crypto/ecdsa"
	"kwil/x/contracts/escrow/evm/abi"
	"kwil/x/crypto"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	client      *ethclient.Client
	ctr         *abi.Escrow
	token       string
	chainId     *big.Int
	privateKey  *ecdsa.PrivateKey
	nodeAddress string
}

func New(client *ethclient.Client, chainId *big.Int, privateKey *ecdsa.PrivateKey, contractAddress string) (*contract, error) {

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

	return &contract{
		client:      client,
		ctr:         ctr,
		token:       tokAddr.Hex(),
		chainId:     chainId,
		privateKey:  privateKey,
		nodeAddress: nodeAddress,
	}, nil
}

func (c *contract) TokenAddress() string {
	return c.token
}
