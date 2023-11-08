package chainsyncer_test

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
	escrowAddr string = "0xcc46cc0960d6903a5b7a76d431aed56fef70e7b0"
	tokenAddr  string = "0x8ce9d23b427b80ab5e21c272a46acd3a27082836"
	userPk     string = "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	userAddr   string = "1e59ce931B4CFea3fe4B875411e280e173cB7A9C"
)

func Test_BridgeClient(t *testing.T) {
	ctx := context.Background()
	bc, err := bClient.New("http://localhost:8545", chain.GOERLI, escrowAddr, tokenAddr)
	assert.NoError(t, err)

	client := bc.ChainClient()
	assert.NotNil(t, client)

	tokenCtr := bc.TokenContract()
	assert.NotNil(t, tokenCtr)

	escrowCtr := bc.EscrowContract()
	assert.NotNil(t, escrowCtr)

	tokenAddress := escrowCtr.TokenAddress()
	fmt.Println("Token Address: ", tokenAddress)
	privKey, err := ec.HexToECDSA(userPk)
	assert.NoError(t, err)

	// Approve:
	hash, err := tokenCtr.Approve(ctx, escrowAddr, big.NewInt(100), privKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Allowance:
	allowance, err := tokenCtr.Allowance(ctx, userAddr, escrowAddr)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), allowance)

	// BalanceOf:
	balance, err := tokenCtr.BalanceOf(ctx, userAddr)
	assert.NoError(t, err)
	fmt.Println("OnChain Balance: ", balance)

	// Deposit:
	depositsPre, err := escrowCtr.Balance(ctx, userAddr)
	assert.NoError(t, err)

	hash, err = escrowCtr.Deposit(ctx, big.NewInt(100), privKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Allowance shld be 0:
	allowance, err = tokenCtr.Allowance(ctx, userAddr, escrowAddr)
	assert.NoError(t, err)
	fmt.Println("Allowance: ", allowance)
	assert.Equal(t, int64(0), allowance.Int64())

	// UserDeposits
	deposits, err := escrowCtr.Balance(ctx, userAddr)
	assert.NoError(t, err)
	var sub = big.NewInt(0)
	assert.Equal(t, big.NewInt(100), sub.Sub(deposits, depositsPre))

	time.Sleep(5 * time.Second)
	events, err := escrowCtr.GetDeposits(ctx, 1, nil)
	assert.NoError(t, err)

	for _, event := range events {
		fmt.Println("Event: ", event)
	}

}
