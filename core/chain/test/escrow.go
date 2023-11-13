package escrow

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	EscrowAbi "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/escrow/abi"
	TokenAbi "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/token/abi"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ec "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

func DeployToken() (string, error) {
	conn, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return "", err
	}

	privKey, err := ec.HexToECDSA("dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5")
	if err != nil {
		return "", err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(5))
	if err != nil {
		return "", err
	}

	addr, _, _, err := TokenAbi.DeployErc20(auth, conn)
	if err != nil {
		return "", err
	}
	fmt.Printf("Token address: %x\n", addr)

	return addr.String(), nil
}

func DeployEscrow(tokenAddr string) (string, error) {
	conn, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return "", err
	}

	privKey, err := ec.HexToECDSA("dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5")
	if err != nil {
		return "", err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(5))
	if err != nil {
		return "", err
	}

	addr, _, _, err := EscrowAbi.DeployEscrow(auth, conn, common.HexToAddress(tokenAddr))
	if err != nil {
		return "", err
	}

	fmt.Printf("Escrow address: %x\n", addr)
	return addr.String(), nil
}

func getAuth(conn *ethclient.Client, privateKey string) *bind.TransactOpts {
	nonce, err := GetNonce(conn, privateKey)
	if err != nil {
		return nil
	}

	privKey, err := ec.HexToECDSA(privateKey)
	if err != nil {
		return nil
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(5))
	if err != nil {
		return nil
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	gas, err := suggestGasPrice(conn)
	if err != nil {
		return nil
	}
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gas
	fmt.Println(gas)
	return auth
}

func ApproveErc20Token(spender string, tokenAddr string, amount *big.Int, privateKey string) (string, error) {
	conn, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return "", err
	}

	auth := getAuth(conn, privateKey)
	token, err := TokenAbi.NewErc20(common.HexToAddress(tokenAddr), conn)
	if err != nil {
		return "", err
	}

	tx, err := token.Approve(auth, common.HexToAddress(spender), amount)
	if err != nil {
		return "", err
	}

	fmt.Printf("Approve tx: %x\n", tx.Hash())
	return tx.Hash().String(), nil
}

func DepositToEscrow(escrowAddr string, amount *big.Int, privateKey string) (string, error) {
	conn, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return "", err
	}

	auth := getAuth(conn, privateKey)

	escrow, err := EscrowAbi.NewEscrow(common.HexToAddress(escrowAddr), conn)
	if err != nil {
		return "", err
	}

	tx, err := escrow.Deposit(auth, amount)
	if err != nil {
		return "", err
	}

	fmt.Printf("Deposit tx: %x\n", tx.Hash())
	return tx.Hash().String(), nil
}

func EscrowBalance(escrowAddr string, pkey string) (*big.Int, error) {
	conn, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return nil, err
	}

	escrow, err := EscrowAbi.NewEscrow(common.HexToAddress(escrowAddr), conn)
	if err != nil {
		return nil, err
	}

	privKey, err := ec.HexToECDSA(pkey)
	if err != nil {
		return nil, err
	}

	balance, err := escrow.Balance(nil, ec.PubkeyToAddress(privKey.PublicKey))
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func GetNonce(conn *ethclient.Client, priv string) (uint64, error) {
	privKey, err := ec.HexToECDSA(priv)
	if err != nil {
		return 0, err
	}
	address := ec.PubkeyToAddress(privKey.PublicKey)

	nonce, err := conn.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func suggestGasPrice(conn *ethclient.Client) (*big.Int, error) {
	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	return gasPrice, nil
}

func EthClient() (*ethclient.Client, error) {
	conn, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func RetrieveEvents(contractAddr string, conn *ethclient.Client, startHeight uint64, endHeight *uint64) error {
	contract := common.HexToAddress(contractAddr)

	filteropts := &bind.FilterOpts{
		Start:   startHeight,
		End:     endHeight,
		Context: nil,
	}
	escrow, err := EscrowAbi.NewEscrow(contract, conn)
	if err != nil {
		return err
	}

	itr, err := escrow.FilterDeposit(filteropts)
	if err != nil {
		return err
	}

	for itr.Next() {
		fmt.Println("Deposit event received from: ", itr.Event.Caller, itr.Event.Amount)
	}

	return nil
}

func SubscribeToEvents(contractAddr string, conn *ethclient.Client, startHeight *uint64, depositChan chan *EscrowAbi.EscrowDeposit) (event.Subscription, error) {
	contract := common.HexToAddress(contractAddr)

	watchOpts := &bind.WatchOpts{
		Start:   startHeight,
		Context: context.Background(),
	}
	escrow, err := EscrowAbi.NewEscrow(contract, conn)
	if err != nil {
		return nil, err
	}

	sub, err := escrow.WatchDeposit(watchOpts, depositChan)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return sub, nil
}

func TokenAddress(contractAddr string, conn *ethclient.Client) string {
	contract := common.HexToAddress(contractAddr)

	escrow, err := EscrowAbi.NewEscrow(contract, conn)
	if err != nil {
		return ""
	}

	tokenAddr, err := escrow.EscrowToken(nil)
	if err != nil {
		return ""
	}

	return strings.ToLower(tokenAddr.String())
}
