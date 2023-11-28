package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	cClient "github.com/kwilteam/kwil-db/core/chain/evm"
	"github.com/kwilteam/kwil-db/core/types/chain"

	EscrowAbi "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/escrow/abi"
	TokenAbi "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/token/abi"
)

/*
	Deployer is a package that contains the deployer interface to connect to the
	chain and deploy the smart contracts.

	ChainRPCURL
	ChainCode
	deployerPrivateKey
	deployerAddress
*/

type EthDeployer struct {
	ChainRPCURL string
	ChainCode   chain.ChainCode

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	Address    common.Address

	connOnce sync.Once
	authOnce sync.Once

	// Deployed Addresses
	EscrowAddress string
	TokenAddress  string

	// Deployed Instances
	deployedToken  *TokenAbi.Erc20
	deployedEscrow *EscrowAbi.Escrow

	// BridgeClient
	client *cClient.EthClient
	auth   *bind.TransactOpts
}

func New(ctx context.Context, chainRPCURL, privateKey string) (*EthDeployer, error) {
	pKey, pubKey := getKeys(privateKey)

	deployer := &EthDeployer{
		ChainRPCURL: chainRPCURL,
		privateKey:  pKey,
		publicKey:   pubKey,
		Address:     crypto.PubkeyToAddress(*pubKey),
	}

	return deployer, nil
}

func (d *EthDeployer) DeployToken(ctx context.Context) (string, error) {
	// Deploy the token contract
	client, err := d.getClient(context.Background(), d.ChainCode, d.ChainRPCURL)
	if err != nil {
		return "", err
	}

	auth, err := d.getAccountAuth(context.Background())
	if err != nil {
		return "", err
	}

	// Deploy the token contract
	addr, _, instance, err := TokenAbi.DeployErc20(auth, client.Backend())
	if err != nil {
		return "", err
	}

	d.TokenAddress = addr.Hex()
	d.deployedToken = instance

	fmt.Println("deployedErc20 =", addr.Hex())
	cAuth := d.getCallAuth(ctx, d.Address.Hex())
	balance, err := instance.BalanceOf(cAuth, d.Address)
	if err != nil {
		return addr.Hex(), err
	}
	fmt.Println("deployer balance =", balance)

	auth, err = d.getAccountAuth(ctx)
	if err != nil {
		return addr.Hex(), err
	}
	// TODO: Is this needed???
	_, err = instance.Erc20Transactor.Approve(auth, addr, big.NewInt(12345*int64(10^18)))
	if err != nil {
		return addr.Hex(), err
	}

	return addr.Hex(), nil
}

func (d *EthDeployer) DeployEscrow(ctx context.Context, tokenAddr string) (string, error) {
	// Deploy the escrow contract
	client, err := d.getClient(context.Background(), d.ChainCode, d.ChainRPCURL)
	if err != nil {
		return "", err
	}

	auth, err := d.getAccountAuth(context.Background())
	if err != nil {
		return "", err
	}

	// Deploy the escrow contract
	addr, _, instance, err := EscrowAbi.DeployEscrow(auth, client.Backend(), common.HexToAddress(tokenAddr))
	if err != nil {
		return "", err
	}

	d.EscrowAddress = addr.Hex()
	d.deployedEscrow = instance

	return d.EscrowAddress, nil
}

func getKeys(_privateKey string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	privateKey, err := crypto.HexToECDSA(_privateKey)
	if err != nil {
		panic(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	return privateKey, publicKeyECDSA
}

func (d *EthDeployer) getClient(ctx context.Context, chainCode chain.ChainCode, endpoint string) (*cClient.EthClient, error) {
	var err error
	d.connOnce.Do(func() {
		d.client, err = cClient.New(ctx, endpoint, chainCode)
	})
	return d.client, err
}

func (d *EthDeployer) getAccountAuth(ctx context.Context) (*bind.TransactOpts, error) {
	var err error
	d.authOnce.Do(func() {
		chainID, err := d.client.ChainID(ctx)
		if err != nil {
			return
		}

		d.auth, err = bind.NewKeyedTransactorWithChainID(d.privateKey, chainID)
		if err != nil {
			return
		}
	})

	//fetch the last use nonce of config
	nonce, err := d.client.GetAccountNonce(ctx, d.Address.Hex())
	if err != nil {
		return d.auth, err
	}

	d.auth.Nonce = big.NewInt(int64(nonce))
	d.auth.Value = big.NewInt(0) // in wei
	return d.auth, err
}

func (d *EthDeployer) getCallAuth(ctx context.Context, from string) *bind.CallOpts {
	auth := bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(from),
		Context: ctx,
	}
	return &auth
}
