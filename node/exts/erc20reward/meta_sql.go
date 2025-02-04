package erc20reward

import (
	"context"
	_ "embed"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/exts/evm-sync/chains"
)

var (
	//go:embed meta_schema.sql
	metaSchema    string
	uuidNamespace = types.MustParseUUID("fc2717ab-e5dd-4f42-bd70-8eac96d0d4c9")
)

// createNewRewardInstance stores information about a pending reward.
// It also creates the first epoch for the reward.
func createNewRewardInstance(ctx context.Context, app *common.App, info *userProvidedData) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}INSERT INTO reward_instances(id, chain_id, escrow_address, distribution_period, synced)
	VALUES (
		$id,
		$chain_id,
		$escrow_address,
		$distribution_period,
		false
	)
	`, map[string]any{
		"id":                  info.ID,
		"chain_id":            info.ChainInfo.ID,
		"escrow_address":      info.EscrowAddress.Bytes(),
		"distribution_period": info.DistributionPeriod,
	}, nil)
}

// createEpoch creates a new epoch for a reward.
// It only stores the epoch's ID, start height, and referenced instance
func createEpoch(ctx context.Context, app *common.App, epoch *PendingEpoch, instanceID *types.UUID) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}INSERT INTO epochs(id, created_at, instance_id)
	VALUES (
		$id,
		$created_at,
		$instance_id
	)`, map[string]any{
		"id":          epoch.ID,
		"created_at":  epoch.StartHeight,
		"instance_id": instanceID,
	}, nil)
}

// finalizeEpoch finalizes an epoch.
// It sets the end height, block hash, and reward root
func finalizeEpoch(ctx context.Context, app *common.App, epochID *types.UUID, endHeight int64, blockHeight []byte, root []byte) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE epochs
	SET ended_at = $ended_at,
		block_hash = $block_hash,
		reward_root = $reward_root
	WHERE id = $id
	`, map[string]any{
		"id":           epochID,
		"finalized_at": endHeight,
		"block_hash":   blockHeight,
		"reward_root":  root,
	}, nil)
}

// confirmEpoch confirms an epoch was received on-chain
//
//lint:ignore U1000 This function is not used yet, but will be used in the future
func confirmEpoch(ctx context.Context, app *common.App, epochID *types.UUID) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE epochs
	SET confirmed = true
	WHERE id = $id
	`, map[string]any{
		"id": epochID,
	}, nil)
}

// setRewardSynced sets a reward as synced.
func setRewardSynced(ctx context.Context, app *common.App, id *types.UUID, kwilBlockhash []byte, syncedAt int64, info *syncedRewardData) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE reward_instances
	SET erc20_address = $erc20_address,
		erc20_decimals = $erc20_decimals,
		synced_at = $synced_at,
		synced = true
	WHERE id = $id
	`, map[string]any{
		"id":             id,
		"erc20_address":  info.Erc20Address.Bytes(),
		"erc20_decimals": info.Erc20Decimals,
		"synced_at":      syncedAt,
	}, nil)
}

// getStoredRewards gets all stored rewards.
func getStoredRewards(ctx context.Context, app *common.App) ([]*rewardExtensionInfo, error) {
	var rewards []*rewardExtensionInfo
	// in the below query, we inner join because an instance must always have an epoch
	err := app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT r.id, r.chain_id, r.escrow_address, r.distribution_period, r.synced, r.active,
		r.erc20_address, r.erc20_decimals, r.synced_at, r.balance, e.id AS epoch_id, e.created_at AS epoch_created_at
	FROM reward_instances r
	JOIN epochs e on r.id = e.instance_id AND e.ended_at IS NULL
	`, nil, func(row *common.Row) error {
		if len(row.Values) != 12 {
			return fmt.Errorf("expected 12 values, got %d", len(row.Values))
		}

		escrowAddr, err := bytesToEthAddress(row.Values[3].([]byte))
		if err != nil {
			return err
		}

		chainID := row.Values[1].(string)
		chainInf, ok := chains.GetChainInfoByID(chainID)
		if !ok {
			return fmt.Errorf("chain %s not found", chainID)
		}

		// initialRewardData should always be not null.
		// syncedRewardData will always be null if synced is false,
		// and not null if synced is true.
		reward := &rewardExtensionInfo{
			userProvidedData: userProvidedData{
				ID:                 row.Values[0].(*types.UUID),
				ChainInfo:          &chainInf,
				EscrowAddress:      escrowAddr,
				DistributionPeriod: row.Values[4].(int64),
			},
			synced: row.Values[5].(bool),
			active: row.Values[6].(bool),
		}

		epochID := row.Values[10].(*types.UUID)
		epochCreatedAt := row.Values[11].(int64)

		reward.currentEpoch = &PendingEpoch{
			ID:          epochID,
			StartHeight: epochCreatedAt,
		}

		if !reward.synced {
			rewards = append(rewards, reward)
			return nil
		}

		erc20Addr, err := bytesToEthAddress(row.Values[7].([]byte))
		if err != nil {
			return err
		}

		reward.syncedRewardData = syncedRewardData{
			Erc20Address:  erc20Addr,
			Erc20Decimals: row.Values[8].(int64),
		}
		reward.syncedAt = row.Values[8].(int64)
		reward.ownedBalance = row.Values[9].(*types.Decimal)

		rewards = append(rewards, reward)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

func bytesToEthAddress(bts []byte) (ethcommon.Address, error) {
	if len(bts) != 20 {
		return ethcommon.Address{}, fmt.Errorf("expected 20 bytes, got %d", len(bts))
	}

	return ethcommon.BytesToAddress(bts), nil
}

// addBalanceToReward adds a balance to the database.
// If it is negative, it will subtract.
// Balances owned by the database can be distributed on chain.
// TODO: I think we should delete this. It wont be needed, as balances should
// either be transferred from a user to the db, or added directly to the user
func addBalanceToReward(ctx context.Context, app *common.App, id *types.UUID, amount *types.Decimal) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE reward_instances
	SET balance = balance + $amount
	WHERE id = $id
	`, map[string]any{
		"id":     id,
		"amount": amount,
	}, nil)
}

// creditBalance credits a balance to a user.
// The rewardId is the ID of the reward instance.
// It if is negative, it will subtract.
// TODO: test negative amounts
func creditBalance(ctx context.Context, app *common.App, rewardId *types.UUID, user ethcommon.Address, amount *types.Decimal) error {
	balanceId := userBalanceID(rewardId, user)
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}INSERT INTO user_balances(id, reward_id, address, balance)
	VALUES ($id, $reward_id, $user, $balance)
	ON CONFLICT (id) DO UPDATE SET balance = user_balances.balance + $balance
	`, map[string]any{
		"id":        balanceId,
		"reward_id": rewardId,
		"user":      user.Bytes(),
		"balance":   amount,
	}, nil)
}

// userBalanceID generates a UUID for a user's balance of a certain instance
func userBalanceID(rewardID *types.UUID, user ethcommon.Address) *types.UUID {
	id := types.NewUUIDV5WithNamespace(*uuidNamespace, append(rewardID.Bytes(), user.Bytes()...))
	return &id
}

// setActiveStatus sets the active status of a reward.
func setActiveStatus(ctx context.Context, app *common.App, id *types.UUID, active bool) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE reward_instances
	SET active = $active
	WHERE id = $id
	`, map[string]any{
		"id":     id,
		"active": active,
	}, nil)
}

// deleteRewardInstance deletes a reward.
func deleteRewardInstance(ctx context.Context, app *common.App, extAlias string) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}DELETE FROM reward_instances
	WHERE extension_alias = $extension_alias
	`, map[string]any{
		"extension_alias": extAlias,
	}, nil)
}

// createSchema creates the schema for the meta extension.
// it should be run exactly once (at genesis)
func createSchema(ctx context.Context, app *common.App) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, metaSchema, nil, nil)
}

// issueReward issues a reward to a user.
func issueReward(ctx context.Context, app *common.App, epochID *types.UUID, user ethcommon.Address, amount *types.Decimal) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE reward_instances
	SET balance = balance - $amount;

	{kwil_erc20_meta}INSERT INTO epoch_rewards(epoch_id, recipient, amount)
	VALUES ($id, $reward_id, $user, $amount)
	ON CONFLICT (id, recipient) DO UPDATE SET amount = epoch_rewards.amount + $amount;
	`, map[string]any{
		"id":     epochID,
		"user":   user.Bytes(),
		"amount": amount,
	}, nil)
}
