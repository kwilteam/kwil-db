package erc20reward

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	kcommon "github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/pg"
)

func Test_scaleUpUint256(t *testing.T) {
	d, err := types.ParseDecimal("11.22")
	require.NoError(t, err)

	t.Run("scale up by 4", func(t *testing.T) {
		nd, err := scaleUpUint256(d, 4)
		require.NoError(t, err)
		require.Equal(t, "112200", nd.String())
		require.Equal(t, 78, int(nd.Precision()))
		require.Equal(t, 0, int(nd.Scale()))
	})

	t.Run("scale up by 0, with decimal", func(t *testing.T) {
		nd, err := scaleUpUint256(d, 0)
		require.NoError(t, err)
		require.Equal(t, "11", nd.String())
		require.Equal(t, 78, int(nd.Precision()))
		require.Equal(t, 0, int(nd.Scale()))
	})

	t.Run("scale up by 0, without decimal", func(t *testing.T) {
		d, err := types.ParseDecimal("1122")
		require.NoError(t, err)
		nd, err := scaleUpUint256(d, 0)
		require.NoError(t, err)
		require.Equal(t, "1122", nd.String())
		require.Equal(t, 78, int(nd.Precision()))
		require.Equal(t, 0, int(nd.Scale()))
	})
}

func Test_scaleDownUint256(t *testing.T) {
	d, err := types.ParseDecimal("112200")
	require.NoError(t, err)

	t.Run("scale down by 4", func(t *testing.T) {
		nd, err := scaleDownUint256(d, 4)
		require.NoError(t, err)
		require.Equal(t, "11.2200", nd.String())
		require.Equal(t, 74, int(nd.Precision()))
		require.Equal(t, 4, int(nd.Scale()))
	})

	t.Run("scale down by 0", func(t *testing.T) {
		nd, err := scaleDownUint256(d, 0)
		require.NoError(t, err)
		require.Equal(t, "112200", nd.String())
		require.Equal(t, 78, int(nd.Precision()))
		require.Equal(t, 0, int(nd.Scale()))
	})
}

func openDB(ctx context.Context) (*pg.DB, error) {
	cfg := &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   "localhost",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild",
				DBName: "kwild",
			},
			MaxConns: 50,
		},
	}
	return pg.NewDB(ctx, cfg)
}

func withTx(t *testing.T, ctx context.Context, db *pg.DB, fn func(app *kcommon.App)) {
	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)

	kApp := &kcommon.App{
		DB: tx,
	}

	fn(kApp)

	err = tx.Commit(ctx)
	require.NoError(t, err)
}

//func TestExt(t *testing.T) {
//	ctx := context.Background()
//
//	db, err := openDB(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer db.Close()
//
//	chainID := int64(11_155_111) // sepolia
//	rewardAddress := "0x55EAC662C9D77cb537DBc9A57C0aDa90eB88132d"
//	safeAddress := "0xbBeaaA74777B1dc14935f5b7E96Bb0ed6DBbD596"
//	nonce := int64(10)
//	safeNonce := int64(10)
//	value := int64(0)
//
//	// From the default Hardhat addresses.
//	//ownerAddress := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
//	ceoAddress := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
//	ceoPK := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
//	ctoAddress := "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
//	engAddress := "0x90F79bf6EB2c4f870365E785982E1f101E93b906"
//
//	ceoPrivateKey, err := crypto.HexToECDSA(ceoPK)
//	require.NoError(t, err)
//
//	namespace := "exttest"
//
//	ext := Erc20RewardExt{
//		contractID:    meta.GenRewardContractID(chainID, rewardAddress),
//		alias:         namespace,
//		ContractAddr:  rewardAddress,
//		SafeAddr:      safeAddress,
//		ChainID:       chainID,
//		Signers:       []string{ceoAddress, ctoAddress, engAddress},
//		Threshold:     1, //
//		ContractNonce: 10,
//		SafeNonce:     10,
//	}
//
//	callCtx := &precompiles.ProcedureContext{
//		TxCtx: &kcommon.TxContext{
//			Ctx: ctx,
//			BlockContext: &kcommon.BlockContext{
//				Height: 100,
//			},
//			Caller: ceoAddress,
//		},
//	}
//
//	tx, err := db.BeginTx(ctx)
//	require.NoError(t, err)
//	err = meta.InitTables(ctx, tx)
//	require.NoError(t, err)
//	err = initTables(ctx, tx, namespace)
//	require.NoError(t, err)
//	err = tx.Commit(ctx)
//	require.NoError(t, err)
//
//	withTx(t, ctx, db, func(app *kcommon.App) {
//		_, err = meta.CreateRewardContract(ctx, app.DB, ext.ChainID, ext.ContractAddress, ext.ContractNonce, ext.Threshold, ext.SafeAddress, ext.SafeNonce, ext.Signers)
//		require.NoError(t, err)
//	})
//
//	// Issue reward, twice. at block 100
//	issueAmount, _ := decimal.NewFromString("11")
//	withTx(t, ctx, db, func(kApp *kcommon.App) {
//		_, err = ext.Call(callCtx, kApp, "issue_reward", []any{engAddress, issueAmount})
//		require.NoError(t, err)
//	})
//	withTx(t, ctx, db, func(kApp *kcommon.App) {
//		callCtx.TxCtx.BlockContext.Height += 1
//		_, err = ext.Call(callCtx, kApp, "issue_reward", []any{engAddress, issueAmount})
//		require.NoError(t, err)
//	})
//
//	// Propose reward batch, at block 150
//
//	callCtx.TxCtx.BlockContext.Height = 150
//	withTx(t, ctx, db, func(kApp *kcommon.App) {
//		_, err = ext.Call(callCtx, kApp, "propose_batch", []any{safeNonce})
//		require.NoError(t, err)
//	})
//
//	// Vote reward batch. need to get original pending reward, then calculate root
//	var latestBatch *EpochReward
//	var pendingReward *PendingReward
//	withTx(t, ctx, db, func(kApp *kcommon.App) {
//		result, err := ext.Call(callCtx, kApp, "list_batches", []any{int64(100), int64(1)})
//		require.NoError(t, err)
//		assert.Len(t, result, 1)
//		batches := result[0].([]*EpochReward)
//		assert.Len(t, batches, 1)
//		latestBatch = batches[0]
//		result, err = ext.Call(callCtx, kApp, "list_rewards", []any{latestBatch.StartHeight, latestBatch.EndHeight})
//		assert.NoError(t, err)
//		assert.Len(t, result, 1)
//		pendingRewards := result[0].([]*PendingReward)
//		assert.Len(t, pendingRewards, 1) // only one, since list_pending will aggregate
//		pendingReward = pendingRewards[0]
//		assert.Equal(t, "22", pendingReward.Amount.String())
//	})
//
//	jsonMtree, rootHash, err := reward.GenRewardMerkleTree([]string{pendingReward.Recipient},
//		[]string{pendingReward.Amount.String()}, rewardAddress, fmt.Sprintf("%d", latestBatch.EndHeight))
//	require.NoError(t, err)
//	fmt.Println("jsonMtree:", jsonMtree)
//	fmt.Printf("rootHash: %s, %d\n", rootHash, latestBatch.TotalRewards.BigInt())
//
//	//// generate signature
//	root, err := hex.DecodeString(rootHash)
//	require.NoError(t, err)
//	safeTxData, err := reward.GenPostRewardTxData(root, latestBatch.TotalRewards.BigInt())
//	require.NoError(t, err)
//	fmt.Println("safeTxData:", hex.EncodeToString(safeTxData))
//	_, safeTxHash, err := reward.GenGnosisSafeTx(rewardAddress, safeAddress, value, safeTxData, chainID, nonce)
//	require.NoError(t, err)
//	fmt.Println("safeTxHash:", hex.EncodeToString(safeTxHash))
//	sig, err := reward.EthGnosisSign(safeTxHash, ceoPrivateKey)
//	require.NoError(t, err)
//	fmt.Println("sig:", hex.EncodeToString(sig))
//	signHash := accounts.TextHash(safeTxHash)
//
//	withTx(t, ctx, db, func(kApp *kcommon.App) {
//		_, err = ext.Call(callCtx, kApp, "vote_batch", []any{signHash, sig})
//		require.NoError(t, err)
//	})
//
//	// list finalized rewards
//	withTx(t, ctx, db, func(kApp *kcommon.App) {
//		result, err := ext.Call(callCtx, kApp, "list_finalized", []any{int64(0), int64(0)})
//		require.NoError(t, err)
//		assert.Len(t, result, 1)
//		finalizedRewards := result[0].([]*FinalizedReward)
//		fr := finalizedRewards[0]
//		fmt.Printf("finalizedReward: %+v\n", fr)
//	})
//}
