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
	}, 1*time.Minute, 5*time.Second)

	time.Sleep(2 * time.Second)
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

	time.Sleep(2 * time.Second)

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

func FundValidatorSpecification(ctx context.Context, t *testing.T, sender DeployerDsl, receiverId ed25519.PrivKey) {
	t.Logf("Executing fund validator specification")

	senderPrivKey, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)
	senderAddr := ec.PubkeyToAddress(senderPrivKey.PublicKey).Bytes()

	// Ensure that the fee for a transfer transaction is as expected.
	var transferPrice = big.NewInt(210_000)
	var transferAmt = big.NewInt(1000000000000000000)

	// Get funds to senders address
	// amount := transferPrice + transferAmt
	amt := big.NewInt(0).Add(transferPrice, transferAmt)
	err = sender.Approve(ctx, senderPrivKey, amt)
	require.NoError(t, err)

	bal1, err := sender.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	// deposit amount into escrow
	err = sender.Deposit(ctx, senderPrivKey, amt)
	require.NoError(t, err)

	// Ensure that the sender account is credited with the amount on kwild
	var bal2 *big.Int
	require.Eventually(t, func() bool {
		bal2, err = sender.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return bal2.Cmp(big.NewInt(0).Add(bal1, amt)) == 0
	}, 5*time.Minute, 5*time.Second) // if receiver voted, they'd get the refund here, at same time as dep

	time.Sleep(2 * time.Second) // wait for receiver to vote if they were too late for resoln

	preValBal, err := sender.AccountBalance(ctx, receiverId)
	require.NoError(t, err)

	// Transfer transferAmt to the Validator
	txHash, err := sender.TransferAmt(ctx, receiverId, transferAmt)
	require.NoError(t, err, "failed to send transfer tx")

	// I expect success
	expectTxSuccess(t, sender, ctx, txHash, defaultTxQueryTimeout)()

	time.Sleep(2 * time.Second) // it reports the old state very briefly, wait a sec

	// Check the validator account balance
	postBal, err := sender.AccountBalance(ctx, receiverId)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0).Add(preValBal, transferAmt), postBal)

	// Check the sender account balance
	postBal, err = sender.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)
	expected := big.NewInt(0).Sub(bal2, big.NewInt(0).Add(transferAmt, transferPrice))
	require.Zero(t, expected.Cmp(postBal), "Incorrect balance in the sender account after the transfer")
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

	time.Sleep(2 * time.Second)

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

	// node0, node1, node2, node3 account balances
	preNodeBalances := make(map[string]*big.Int)
	for key := range addresses {
		balance, err := deployer.AccountBalance(ctx, addresses[key])
		require.NoError(t, err)
		preNodeBalances[key] = balance
	}

	// approve 10 tokens
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// I deposit amount into the escrow
	err = deployer.Deposit(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// can we get the block number? The event should have expired by block height = 3/4
	// only 1 node hears and votes on this event.
	time.Sleep(30 * time.Second)

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

func DepositResolutionExpiryRefundSpecification(ctx context.Context, t *testing.T, deployer DeployerDsl, privKeys map[string]ed25519.PrivKey) {
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

	// node0, node1, node2, node3 account balances
	preNodeBalances := make(map[string]*big.Int)
	for key := range addresses {
		balance, err := deployer.AccountBalance(ctx, addresses[key])
		require.NoError(t, err)
		preNodeBalances[key] = balance
	}

	// approve 10 tokens
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// I deposit amount into the escrow
	err = deployer.Deposit(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// can we get the block number? The event should have expired by block height = 3/4
	time.Sleep(30 * time.Second)

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
		// Check that the node0,1 issued a vote but got refunded as minthreshold for expiry refund met.
		// Check that the node2,3,4 didn't issue any Votes, thus no tx fee lost
		require.Equal(t, preNodeBalances[key], postNodeBalances[key])
	}
}

func EthDepositValidatorUpdatesSpecification(ctx context.Context, t *testing.T, valDsl map[string]ValidatorOpsDsl, deployer DeployerDsl, privKeys map[string]ed25519.PrivKey) {
	t.Logf("Executing validator updates specification")

	sender, err := ec.HexToECDSA(senderPk)
	require.NoError(t, err)
	senderAddr := ec.PubkeyToAddress(sender.PublicKey).Bytes()

	// Start the test with 6 nodes with eth deposit oracle enabled
	// out of which 2 nodes are in byzantine mode [node0, node1]
	// Make a Deposit and ensure that its credited to the account

	// get user balance
	bal1, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)

	// approve 10 tokens
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// I deposit amount into the escrow
	err = deployer.Deposit(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	var bal2 *big.Int
	// Check that the user balance is updated
	require.Eventually(t, func() bool {
		bal2, err = deployer.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return bal2.Cmp(big.NewInt(0).Add(bal1, big.NewInt(10))) == 0
	}, 5*time.Minute, 5*time.Second)

	// Node5 leaves its validator status
	ValidatorNodeLeaveSpecification(ctx, t, valDsl["node5"])

	// Check that the node5 is no more a Validator
	CurrentValidatorsSpecification(ctx, t, valDsl["node5"], 5)

	// Make a deposit
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	err = deployer.Deposit(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	// total 5 validators and 3 listening to the deposit event
	// 5 * 2/3 = 3.33 => required 4 validators to vote for deposit to get credited
	// Check that the user balance is not updated
	time.Sleep(15 * time.Second)
	bal3, err := deployer.AccountBalance(ctx, senderAddr)
	require.NoError(t, err)
	require.Equal(t, bal2, bal3)

	// Node5 rejoins the network as a validator
	// And catches up with all the events it missed and votes for the observed events
	// The last deposit should now get approved and credited to the account
	joinerPubKey := privKeys["node5"].PubKey().Bytes()
	ValidatorNodeJoinSpecification(ctx, t, valDsl["node5"], joinerPubKey, 5)
	// needs 4 approvals
	for i := 0; i < 3; i++ {
		node := fmt.Sprintf("node%d", i)
		ValidatorNodeApproveSpecification(ctx, t, valDsl[node], joinerPubKey, 5, 5, false)
	}
	ValidatorNodeApproveSpecification(ctx, t, valDsl["node3"], joinerPubKey, 5, 6, true)

	// Check that the node5 became a Validator
	CurrentValidatorsSpecification(ctx, t, valDsl["node5"], 6)

	// Ensure that the previous unapproved deposits are now approved
	// Check that the user balance is updated
	var bal4 *big.Int
	require.Eventually(t, func() bool {
		bal4, err = deployer.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return bal4.Cmp(big.NewInt(0).Add(bal3, big.NewInt(10))) == 0
	}, 5*time.Minute, 5*time.Second)

	// Make one more deposit and ensure that it gets credited
	err = deployer.Approve(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	err = deployer.Deposit(ctx, sender, big.NewInt(10))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		bal5, err := deployer.AccountBalance(ctx, senderAddr)
		require.NoError(t, err)
		return bal5.Cmp(big.NewInt(0).Add(bal4, big.NewInt(10))) == 0
	}, 5*time.Minute, 5*time.Second)
}
