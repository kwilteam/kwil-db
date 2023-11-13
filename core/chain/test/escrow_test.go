package escrow

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	tokenAddr  string = "0x274b028b03a250ca03644e6c578d81f019ee1323"
	escrowAddr string = "0xbcf7fffd8b256ec51a36782a52d0c34f6474d951"
)

func Test_DeployEscrowToken(t *testing.T) {
	tokenAddr, err := DeployToken()
	assert.NoError(t, err) // 274b028b03a250ca03644e6c578d81f019ee1323

	escrowAddr, err := DeployEscrow(tokenAddr) // bcf7fffd8b256ec51a36782a52d0c34f6474d951
	assert.NoError(t, err)

	assert.NotEmpty(t, tokenAddr)
	assert.NotEmpty(t, escrowAddr)
}

func Test_EscrowToken(t *testing.T) {
	conn, err := EthClient()
	assert.NoError(t, err)

	TokenAddress := TokenAddress(escrowAddr, conn)
	assert.Equal(t, tokenAddr, TokenAddress)

}

func Test_PendingNonce(t *testing.T) {
	conn, err := EthClient()
	assert.NoError(t, err)

	nonce, err := GetNonce(conn, "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5")
	assert.NoError(t, err)
	fmt.Println(nonce, "   Error: ", err)

	nonce, err = GetNonce(conn, "16dd30d52297ff9973cbbd5f35c0fef37309fbbfd5b540615b255fbeb8c1283d")
	assert.NoError(t, err)
	fmt.Println(nonce, "   Error: ", err)
}

func Test_ApproveErc20Token(t *testing.T) {
	// spender is the escrow contract address
	spender := escrowAddr

	// private key of the token owner
	privKey := "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"

	hash, err := ApproveErc20Token(spender, tokenAddr, big.NewInt(100000), privKey)
	assert.NoError(t, err)
	fmt.Println(hash)

	conn, err := EthClient()
	assert.NoError(t, err)
	ctx := context.Background()
	_, pending, err := conn.TransactionByHash(ctx, common.HexToHash(hash))
	assert.NoError(t, err)
	fmt.Println(pending)
}

func Test_Deposits(t *testing.T) {
	pKey1 := "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	pkey2 := "16dd30d52297ff9973cbbd5f35c0fef37309fbbfd5b540615b255fbeb8c1283d"

	hash1, err := DepositToEscrow(escrowAddr, big.NewInt(1000), pKey1)
	assert.NoError(t, err)

	hash2, err := DepositToEscrow(escrowAddr, big.NewInt(1000), pkey2)
	assert.NoError(t, err)

	// Check if transactions are still pending
	conn, err := EthClient()
	assert.NoError(t, err)
	ctx := context.Background()
	_, pending, err := conn.TransactionByHash(ctx, common.HexToHash(hash1))
	assert.NoError(t, err)
	assert.False(t, pending)

	_, pending, err = conn.TransactionByHash(ctx, common.HexToHash(hash2))
	assert.NoError(t, err)
	assert.False(t, pending)

	receipt, err := conn.TransactionReceipt(ctx, common.HexToHash(hash1))
	assert.NoError(t, err)
	fmt.Println(receipt.Status)
}

func Test_Balance(t *testing.T) {
	pKey1 := "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	pkey2 := "16dd30d52297ff9973cbbd5f35c0fef37309fbbfd5b540615b255fbeb8c1283d"

	balance, err := EscrowBalance(escrowAddr, pKey1)
	assert.NoError(t, err)
	fmt.Println(balance)

	balance, err = EscrowBalance(escrowAddr, pkey2)
	assert.NoError(t, err)
	fmt.Println(balance)
}

func Test_DepositEvents(t *testing.T) {
	conn, err := EthClient()
	assert.NoError(t, err)
	//endHeight := uint64(80)
	RetrieveEvents(escrowAddr, conn, 0, nil)
}

func Test_Deposit_Subscription(t *testing.T) {
	// conn, err := EthClient()
	// assert.NoError(t, err)
	// escrowAddr := "0xcc46cc0960d6903a5b7a76d431aed56fef70e7b0"
	// startHeight := uint64(0)

	// depositEvents := make(chan *EscrowAbi.EscrowDeposit)
	// subscription, err := SubscribeToEvents(escrowAddr, conn, &startHeight, depositEvents)
	// assert.NoError(t, err)
	// assert.NotNil(t, subscription)

	// //defer subscription.Unsubscribe()

	// go func(events chan *EscrowAbi.EscrowDeposit, subscription event.Subscription) {
	// 	for {
	// 		select {
	// 		case err := <-subscription.Err():
	// 			fmt.Println(err)
	// 		case event := <-events:
	// 			fmt.Println("Deposit event received from: ", event.Caller, event.Amount)
	// 		}
	// 	}
	// }(depositEvents, subscription)

	// // Wait for events
	// Test_Deposits(t)

	// // Wait for events
	// time.Sleep(10 * time.Second)
}
