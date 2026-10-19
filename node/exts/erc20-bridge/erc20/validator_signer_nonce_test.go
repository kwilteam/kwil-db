package erc20

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/trufnetwork/kwil-db/common"
	"github.com/trufnetwork/kwil-db/config"
	kwilcrypto "github.com/trufnetwork/kwil-db/core/crypto"
	"github.com/trufnetwork/kwil-db/core/crypto/auth"
	"github.com/trufnetwork/kwil-db/core/log"
	"github.com/trufnetwork/kwil-db/core/types"
	"github.com/trufnetwork/kwil-db/node/types/sql"
)

// nonceTestAccounts reports a nonce the test moves by hand, the way the node's
// other transactions move the validator account's nonce between votes.
type nonceTestAccounts struct {
	common.Accounts // only GetAccount is called
	nonce           int64
}

func (a *nonceTestAccounts) GetAccount(_ context.Context, _ sql.Executor, id *types.AccountID) (*types.Account, error) {
	return &types.Account{ID: id, Balance: big.NewInt(0), Nonce: a.nonce}, nil
}

type nonceTestReadTx struct{ sql.OuterReadTx }

func (nonceTestReadTx) Rollback(context.Context) error { return nil }

type nonceTestReadTxMaker struct{}

func (nonceTestReadTxMaker) BeginDelayedReadTx() sql.OuterReadTx { return nonceTestReadTx{} }

// nonceTestKey signs the vote transaction with a plain key.
type nonceTestKey struct {
	common.ValidatorSigner // only CreateSecp256k1Signer is called
	key                    *ecdsa.PrivateKey
}

func (k nonceTestKey) CreateSecp256k1Signer() (auth.Signer, error) {
	priv, err := kwilcrypto.UnmarshalSecp256k1PrivateKey(crypto.FromECDSA(k.key))
	if err != nil {
		return nil, err
	}
	return &auth.EthPersonalSigner{Key: *priv}, nil
}

// The vote's nonce has to follow the validator's account, not a count kept since
// the first vote. The node signs other transactions with the same key, so a
// count of its own votes falls behind, and every vote after that is rejected as
// an invalid nonce until the count catches up.
func TestVoteTransactionNonceFollowsTheAccount(t *testing.T) {
	ctx := context.Background()
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	accounts := &nonceTestAccounts{nonce: 11750}
	v := &ValidatorSigner{
		app: &common.App{
			Accounts: accounts,
			Service: &common.Service{
				Logger:        log.DiscardLogger,
				GenesisConfig: &config.GenesisConfig{ChainID: "nonce-test"},
			},
		},
		dbPool:          nonceTestReadTxMaker{},
		instanceID:      types.NewUUIDV5([]byte("nonce-test-instance")),
		validatorSigner: nonceTestKey{key: key},
		address:         crypto.PubkeyToAddress(key.PublicKey),
		logger:          log.DiscardLogger,
		votedEpochs:     make(map[string]bool),
	}
	epochID := types.NewUUIDV5([]byte("nonce-test-epoch"))

	voteNonce := func() uint64 {
		t.Helper()
		tx, err := v.createVoteTransaction(ctx, epochID, make([]byte, 65))
		require.NoError(t, err)
		return tx.Body.Nonce
	}

	// The account's nonce is the last one it used, so the first vote after the
	// node starts takes the next one.
	require.Equal(t, uint64(11751), voteNonce())

	// Other transactions from the same key move the account, and the next vote
	// follows it.
	accounts.nonce += 328
	require.Equal(t, uint64(12079), voteNonce())

	// A vote the mempool turned away leaves nothing behind: until the account
	// moves, a retry asks for the same nonce.
	require.Equal(t, uint64(12079), voteNonce())

	// A key that has never sent a transaction starts at nonce 1.
	accounts.nonce = 0
	require.Equal(t, uint64(1), voteNonce())
}
