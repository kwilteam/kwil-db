package tokenbridge_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types/chain"
	bClient "github.com/kwilteam/kwil-db/oracles/tokenbridge/client"
	"github.com/stretchr/testify/assert"

	ec "github.com/ethereum/go-ethereum/crypto"
)

// TODO: remove this file

var (
	escrowAddr string = "0xbcf7fffd8b256ec51a36782a52d0c34f6474d951"
	//tokenAddr  string = "0xccf612a958da1f8d3fa97a447fc44cffe9994a54"
	userPk   string = "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	userAddr string = "1e59ce931B4CFea3fe4B875411e280e173cB7A9C"

	// userPk2 string = "6789ede33b84cbd4e735e12924d07e48b15df0ded10de3c206eeac585852ab22"
	// userAddr2 string = "c89D42189f0450C2b2c3c61f58Ec5d628176A1E7"
)

func TestDeposits(t *testing.T) {
	ctx := context.Background()
	bc, err := bClient.New(ctx, "http://localhost:8545", chain.GOERLI, escrowAddr)
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
	bc, err := bClient.New(ctx, "http://localhost:8545", chain.GOERLI, escrowAddr)
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

func Test_Marshalling(t *testing.T) {
	type config1 struct {
		Param1 string `json:"param1"`
		Param2 string `json:"param2"`
	}

	type config2 struct {
		Value1 string `json:"param1"`
		Value2 string `json:"param2"`
	}

	c1 := &config1{
		Param1: "paramval1",
		Param2: "paramval2",
	}
	bts, err := json.Marshal(c1)
	assert.NoError(t, err)

	c2 := &config2{}
	err = json.Unmarshal(bts, c2)
	assert.NoError(t, err)
	fmt.Println("c2: ", c2)
	assert.Equal(t, c1.Param1, c2.Value1)
	assert.Equal(t, c1.Param2, c2.Value2)
}
