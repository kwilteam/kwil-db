package tokenbridge_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	bClient "github.com/kwilteam/kwil-db/core/bridge/client"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/stretchr/testify/assert"

	ec "github.com/ethereum/go-ethereum/crypto"
)

var (
	escrowAddr string = "0x006c992966be10e1da52fb2b09a62a1059c093bf"
	//tokenAddr  string = "0xccf612a958da1f8d3fa97a447fc44cffe9994a54"
	userPk   string = "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	userAddr string = "1e59ce931B4CFea3fe4B875411e280e173cB7A9C"
)

func TestDeposits(t *testing.T) {
	ctx := context.Background()
	bc, err := bClient.New("http://localhost:8545", chain.GOERLI, escrowAddr)
	assert.NoError(t, err)

	privKey, _ := ec.HexToECDSA(userPk)
	// Approve:
	hash, err := bc.Approve(ctx, escrowAddr, big.NewInt(100), privKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	hash, err = bc.Deposit(ctx, big.NewInt(100), privKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func Test_BridgeClient(t *testing.T) {
	ctx := context.Background()
	// bc, err := bClient.New("http://localhost:8545", chain.GOERLI, escrowAddr, tokenAddr)
	bc, err := bClient.New("http://localhost:8545", chain.GOERLI, escrowAddr)
	assert.NoError(t, err)

	tokenAddress := bc.TokenAddress()
	fmt.Println("Token Address: ", tokenAddress)
	privKey, err := ec.HexToECDSA(userPk)
	assert.NoError(t, err)

	// Approve:
	hash, err := bc.Approve(ctx, escrowAddr, big.NewInt(100), privKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Allowance:
	allowance, err := bc.Allowance(ctx, userAddr, escrowAddr)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), allowance)

	// BalanceOf:
	balance, err := bc.BalanceOf(ctx, userAddr)
	assert.NoError(t, err)
	fmt.Println("OnChain Balance: ", balance)

	// Deposit:
	depositsPre, err := bc.DepositBalance(ctx, userAddr)
	assert.NoError(t, err)

	hash, err = bc.Deposit(ctx, big.NewInt(100), privKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Allowance shld be 0:
	allowance, err = bc.Allowance(ctx, userAddr, escrowAddr)
	assert.NoError(t, err)
	fmt.Println("Allowance: ", allowance)
	assert.Equal(t, int64(0), allowance.Int64())

	// UserDeposits
	deposits, err := bc.DepositBalance(ctx, userAddr)
	assert.NoError(t, err)
	var sub = big.NewInt(0)
	assert.Equal(t, big.NewInt(100), sub.Sub(deposits, depositsPre))

	time.Sleep(5 * time.Second)
	events, err := bc.GetDeposits(ctx, 1, nil)
	assert.NoError(t, err)

	for _, event := range events {
		fmt.Println("Event: ", event)
	}
}
