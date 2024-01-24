package specifications

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	ec "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

const (
	senderPk = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
)

var (
	deployPrice = big.NewInt(1000000000000000000)
)

func ApproveSpecification(ctx context.Context, t *testing.T, deployer DeployerDsl) {
	t.Logf("Executing approve specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)

	// I approve 20 tokens
	err = deployer.Approve(ctx, sender, big.NewInt(20))
	require.NoError(t, err)

	// I expect the allowance to be 20
	allowance, err := deployer.Allowance(ctx, sender)
	require.NoError(t, err)

	// I expect the allowance to be 20
	require.Equal(t, big.NewInt(20), allowance)

	// I approve 10 tokens
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// I expect the allowance to be 10
	allowance, err = deployer.Allowance(ctx, sender)
	require.NoError(t, err)

	// I expect the allowance to be 10
	require.Equal(t, big.NewInt(10), allowance)

}

func DepositSuccessSpecification(ctx context.Context, t *testing.T, deployer DeployerDsl, amount *big.Int) {
	t.Logf("Executing deposit specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)

	// I approve 10 tokens
	err = deployer.Approve(ctx, sender, amount)
	require.NoError(t, err)

	// I expect the allowance to be 10
	allowance, err := deployer.Allowance(ctx, sender)
	require.NoError(t, err)
	require.Equal(t, amount, allowance)

	// Get the escrow balance
	preBalance, err := deployer.EscrowBalance(ctx, sender)
	require.NoError(t, err)

	senderAddr := ec.PubkeyToAddress(sender.PublicKey).Bytes()
	preUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	// I deposit amount into the escrow
	err = deployer.Deposit(ctx, sender, amount)
	require.NoError(t, err)

	// Get the escrow balance
	balance, err := deployer.EscrowBalance(ctx, sender)
	require.NoError(t, err)

	// I expect the balance to be equal to amount
	require.Equal(t, amount, big.NewInt(0).Sub(balance, preBalance))

	require.Eventually(t, func() bool {
		// Want postUserBalance to be equal to amount + preUserBalance
		postUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return postUserBalance.Cmp(big.NewInt(0).Add(preUserBalance, amount)) == 0
	}, 5*time.Minute, 5*time.Second)
}

func DepositFailSpecification(ctx context.Context, t *testing.T, deployer DeployerDsl) {
	t.Logf("Executing deposit fail specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)

	// I approve 10 tokens
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// I expect the allowance to be 10
	allowance, err := deployer.Allowance(ctx, sender)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(10), allowance)

	// Get the escrow balance
	preBalance, err := deployer.EscrowBalance(ctx, sender)
	require.NoError(t, err)

	// I deposit 20 tokens into the escrow
	err = deployer.Deposit(ctx, sender, big.NewInt(20))
	require.NoError(t, err)

	// Get the escrow balance
	postBalance, err := deployer.EscrowBalance(ctx, sender)
	require.NoError(t, err)

	// 20 tokens should not be deposited, balance should be same as before
	require.Equal(t, preBalance, postBalance)
}

func DeployDbInsufficientFundsSpecification(ctx context.Context, t *testing.T, deployer DeployerDsl) {
	t.Logf("Executing deploy db insufficient funds specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)
	amount := big.NewInt(10)

	// I approve 10 tokens
	err = deployer.Approve(ctx, sender, amount)
	require.NoError(t, err)

	senderAddr := ec.PubkeyToAddress(sender.PublicKey).Bytes()

	preUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	// I deposit amount into the escrow
	err = deployer.Deposit(ctx, sender, amount)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		// Want postUserBalance to be equal to amount + preUserBalance
		postUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return postUserBalance.Cmp(big.NewInt(0).Add(preUserBalance, amount)) == 0
	}, 5*time.Minute, 5*time.Second)

	// Should be able to deploy database
	db := SchemaLoader.Load(t, SchemaTestDB)

	// When i deploy the database
	txHash, err := deployer.DeployDatabase(ctx, db)
	require.NoError(t, err, "failed to send deploy database tx")

	// Then i expect success
	expectTxFail(t, deployer, ctx, txHash, defaultTxQueryTimeout)()

	// And i expect database should exist
	err = deployer.DatabaseExists(ctx, deployer.DBID(db.Name))
	require.Error(t, err)

	// Check the user balance
	postDeployBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	// User balance reduced to 0, as it submitted a deploy request without sufficient funds
	require.Equal(t, big.NewInt(0), postDeployBalance)
}

func DeployDbSuccessSpecification(ctx context.Context, t *testing.T, deployer DeployerDsl) {
	t.Logf("Executing deposit deploy db success specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)
	senderAddr := ec.PubkeyToAddress(sender.PublicKey).Bytes()

	// I approve 10 tokens
	amount := big.NewInt(0).Add(deployPrice, big.NewInt(10))
	err = deployer.Approve(ctx, sender, amount)
	require.NoError(t, err)

	preUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	// I deposit amount into the escrow
	err = deployer.Deposit(ctx, sender, amount)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		// Want postUserBalance to be equal to amount + preUserBalance
		postUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return postUserBalance.Cmp(big.NewInt(0).Add(preUserBalance, amount)) == 0
	}, 5*time.Minute, 5*time.Second)

	// Should be able to deploy database
	db := SchemaLoader.Load(t, SchemaTestDB)

	// When i deploy the database
	txHash, err := deployer.DeployDatabase(ctx, db)
	require.NoError(t, err, "failed to send deploy database tx")

	// Then i expect success
	expectTxSuccess(t, deployer, ctx, txHash, defaultTxQueryTimeout)()

	// And i expect database should exist
	err = deployer.DatabaseExists(ctx, deployer.DBID(db.Name))
	require.NoError(t, err)

	// Check the user balance
	postDeployBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	var diff = big.NewInt(0)
	// User balance reduced to 0, as it submitted a deploy request without sufficient funds
	require.Equal(t, diff.Sub(postDeployBalance, preUserBalance), big.NewInt(10))
}

// Can only be run in Byzantine mode.
func DepositResolutionExpirySpecification(ctx context.Context, t *testing.T, deployer DeployerDsl, privKeys map[string]ed25519.PrivKey) {
	t.Logf("Executing deposit resolution expiry specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)
	senderAddr := ec.PubkeyToAddress(sender.PublicKey).Bytes()

	addresses := make(map[string][]byte)
	for key := range privKeys {
		addresses[key] = privKeys[key].PubKey().Bytes()
	}

	// get user balance
	preUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)
	fmt.Println("User Balance: ", preUserBalance.String())

	// node0, node1, node2, node3 account balances
	preNodeBalances := make(map[string]*big.Int)
	for key := range addresses {
		balance, err := deployer.AccountBalance(ctx, addresses[key])
		require.NoError(t, err)
		preNodeBalances[key] = balance
	}

	// can we get the block number? The event should have expired by block height = 3/4
	time.Sleep(20 * time.Second)

	// Check that the user balance is not updated
	postUserBalance, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)
	require.Equal(t, preUserBalance, postUserBalance)

	postNodeBalances := make(map[string]*big.Int)
	for key := range addresses {
		balance, err := deployer.AccountBalance(ctx, addresses[key])
		require.NoError(t, err)
		postNodeBalances[key] = balance
	}

	for key := range addresses {
		// Check that the node0 issued a vote and lost the tx fee
		if key == "node0" {
			require.Less(t, postNodeBalances[key].Int64(), preNodeBalances[key].Int64())
		} else {
			// Check that the node1,2,3 didn't issue any Votes, thus no tx fee lost
			require.Equal(t, preNodeBalances[key], postNodeBalances[key])
		}
	}
}
