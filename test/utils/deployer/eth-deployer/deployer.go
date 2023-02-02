package eth_deployer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"kwil/pkg/fund"
	escrow "kwil/x/contracts/escrow/evm/abi"
	token "kwil/x/contracts/token/evm/abi"
	"math/big"
	"sync"
)

const (
	// TotalSupply test token total supply is TotalSupply*10^18
	// change x/contracts/token/evm/abi/erc20.bin if you want to change it
	TotalSupply = 12345
)

type EthDeployer struct {
	Provider string
	PriKey   string

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey

	auth      *bind.TransactOpts
	Account   common.Address
	authOnce  sync.Once
	connOnce  sync.Once
	ethClient *ethclient.Client

	deployedEscrow *escrow.Escrow
	deployedErc20  *token.Erc20

	Chain fund.IFund

	domination *big.Int
}

func NewEthDeployer(provider string, _privateKey string, domination *big.Int) *EthDeployer {
	privateKey, publicKey := getKeys(_privateKey)

	d := &EthDeployer{
		Provider:   provider,
		PriKey:     _privateKey,
		privateKey: privateKey,
		publicKey:  publicKey,
		Account:    crypto.PubkeyToAddress(*publicKey),
		domination: domination,
	}

	return d
}

func (d *EthDeployer) UpdateContract(ctx context.Context, poolAddress string) error {
	client, err := d.getClient()
	if err != nil {
		return err
	}

	escrowCtl, err := escrow.NewEscrow(common.HexToAddress(poolAddress), client)
	if err != nil {
		return fmt.Errorf("failed to create escrow contract: %v", err)
	}

	cAuth := d.getCallAuth(ctx, d.Account.Hex())
	tokenAddress, err := escrowCtl.EscrowToken(cAuth)
	if err != nil {
		return err
	}

	tokenCtl, err := token.NewErc20(tokenAddress, client)
	if err != nil {
		return fmt.Errorf("failed to create erc20 contract: %v", err)
	}

	d.deployedEscrow = escrowCtl
	d.deployedErc20 = tokenCtl
	return nil
}

func (d *EthDeployer) GetPrivateKey() *ecdsa.PrivateKey {
	return d.privateKey
}

func (d *EthDeployer) getClient() (*ethclient.Client, error) {
	var err error
	d.connOnce.Do(func() {
		d.ethClient, err = ethclient.Dial(d.Provider)
	})
	return d.ethClient, err
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

func (d *EthDeployer) DeployEscrow(ctx context.Context, tokenAddr string) (common.Address, error) {
	var deployedAddr common.Address

	client, err := d.getClient()
	if err != nil {
		return deployedAddr, err
	}

	auth, err := d.getAccountAuth(ctx)
	if err != nil {
		return deployedAddr, err
	}

	deployedAddr, _, instance, err := escrow.DeployEscrow(auth, client, common.HexToAddress(tokenAddr))
	if err != nil {
		return deployedAddr, err
	}

	d.deployedEscrow = instance

	fmt.Println("deployedEscrow=", deployedAddr.Hex())

	//cAuth := d.getCallAuth(client, d.Account.Hex())
	//tokenAddress, err := instance.EscrowToken(cAuth)
	//if err != nil {
	//	return deployedAddr, err
	//}
	//fmt.Println("escrowTokenAddress=", tokenAddress.Hex())

	return deployedAddr, nil
}

func (d *EthDeployer) DeployToken(ctx context.Context) (common.Address, error) {
	var deployedAddr common.Address

	client, err := d.getClient()
	if err != nil {
		return deployedAddr, err
	}
	auth, err := d.getAccountAuth(ctx)
	if err != nil {
		return deployedAddr, err
	}

	deployedAddr, _, instance, err := token.DeployErc20(auth, client)
	if err != nil {
		return deployedAddr, err
	}

	d.deployedErc20 = instance

	fmt.Println("deployedErc20=", deployedAddr.Hex())
	cAuth := d.getCallAuth(ctx, d.Account.Hex())
	balance, err := instance.BalanceOf(cAuth, d.Account)
	if err != nil {
		return deployedAddr, err
	}
	fmt.Println("deployer balance =", balance)

	auth, err = d.getAccountAuth(ctx)
	if err != nil {
		return deployedAddr, err
	}
	instance.Erc20Transactor.Approve(auth, deployedAddr, big.NewInt(TotalSupply*int64(10^18)))

	return deployedAddr, nil
}

func (d *EthDeployer) FundAccount(ctx context.Context, account string, amount int64) error {
	// transfer eth to info
	_, err := d.getClient()
	if err != nil {
		return err
	}

	// transfer erc20 token to info
	//cAuth := d.getCallAuth(ctx, d.Account.Hex())
	//decimals, err := d.deployedErc20.Decimals(cAuth)
	//fmt.Println("token decimals = ", decimals)
	realAmount := new(big.Int).Mul(big.NewInt(amount), d.domination)

	auth, err := d.getAccountAuth(ctx)
	if err != nil {
		return err
	}
	_, err = d.deployedErc20.Erc20Transactor.Transfer(auth, common.HexToAddress(account), realAmount)
	if err != nil {
		return err
	}

	//balance, err := d.deployedErc20.BalanceOf(cAuth, common.HexToAddress(info))
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("info(%s) balance = %d\n", info, balance)
	return err
}

func (d *EthDeployer) getCallAuth(ctx context.Context, from string) *bind.CallOpts {
	auth := bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(from),
		Context: ctx,
	}
	return &auth
}

func (d *EthDeployer) getAccountAuth(ctx context.Context) (*bind.TransactOpts, error) {
	var err error
	d.authOnce.Do(func() {
		chainID, err := d.ethClient.ChainID(ctx)
		if err != nil {
			return
		}

		d.auth, err = bind.NewKeyedTransactorWithChainID(d.privateKey, chainID)
		if err != nil {
			return
		}
	})

	//fetch the last use nonce of info
	nonce, err := d.ethClient.PendingNonceAt(ctx, d.Account)
	if err != nil {
		return d.auth, err
	}

	d.auth.Nonce = big.NewInt(int64(nonce))
	d.auth.Value = big.NewInt(0) // in wei
	//auth.GasLimit = uint64(3000000) // in units
	//auth.GasPrice = big.NewInt(1000000)

	return d.auth, err
}
