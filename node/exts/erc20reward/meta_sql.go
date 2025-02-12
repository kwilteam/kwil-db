package erc20reward

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/engine"
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
	{kwil_erc20_meta}INSERT INTO epochs(id, created_at_block, created_at_unix, instance_id)
	VALUES (
		$id,
		$created_at_block,
		$created_at_unix,
		$instance_id
	)`, map[string]any{
		"id":               epoch.ID,
		"created_at_block": epoch.StartHeight,
		"created_at_unix":  epoch.StartTime.Unix(),
		"instance_id":      instanceID,
	}, nil)
}

// finalizeEpoch finalizes an epoch.
// It sets the end height, block hash, and reward root
func finalizeEpoch(ctx context.Context, app *common.App, epochID *types.UUID, endHeight int64, blockHash []byte, root []byte) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE epochs
	SET ended_at = $ended_at,
		block_hash = $block_hash,
		reward_root = $reward_root
	WHERE id = $id
	`, map[string]any{
		"id":          epochID,
		"ended_at":    endHeight,
		"block_hash":  blockHash,
		"reward_root": root,
	}, nil)
}

// confirmEpoch confirms an epoch was received on-chain
func confirmEpoch(ctx context.Context, app *common.App, root []byte) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE epochs
	SET confirmed = true
	WHERE reward_root = $root
	`, map[string]any{
		"root": root,
	}, nil)
}

// setRewardSynced sets a reward as synced.
func setRewardSynced(ctx context.Context, app *common.App, id *types.UUID, syncedAt int64, info *syncedRewardData) error {
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

// getStoredRewardInstances gets all stored reward instances.
func getStoredRewardInstances(ctx context.Context, app *common.App) ([]*rewardExtensionInfo, error) {
	var rewards []*rewardExtensionInfo
	err := app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT r.id, r.chain_id, r.escrow_address, r.distribution_period, r.synced, r.active,
		r.erc20_address, r.erc20_decimals, r.synced_at, r.balance, e.id AS epoch_id,
		e.created_at_block AS epoch_created_at_block, e.created_at_unix AS epoch_created_at_seconds
	FROM reward_instances r
	LEFT JOIN epochs e on r.id = e.instance_id AND e.confirmed IS NULL
	`, nil, func(row *common.Row) error {
		if len(row.Values) != 13 {
			return fmt.Errorf("expected 13 values, got %d", len(row.Values))
		}

		escrowAddr, err := bytesToEthAddress(row.Values[2].([]byte))
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
				DistributionPeriod: row.Values[3].(int64),
			},
			synced: row.Values[4].(bool),
			active: row.Values[5].(bool),
		}

		if row.Values[10] == nil {
			return fmt.Errorf("internal bug: instance %s has no epoch", reward.ID)
		}

		epochID := row.Values[10].(*types.UUID)
		epochCreatedAtBlock := row.Values[11].(int64)
		epochCreatedAtUnix := row.Values[12].(int64)

		reward.currentEpoch = &PendingEpoch{
			ID:          epochID,
			StartHeight: epochCreatedAtBlock,
			StartTime:   time.Unix(epochCreatedAtUnix, 0),
		}

		if !reward.synced {
			rewards = append(rewards, reward)
			return nil
		}

		erc20Addr, err := bytesToEthAddress(row.Values[6].([]byte))
		if err != nil {
			return err
		}

		reward.syncedRewardData = syncedRewardData{
			Erc20Address:  erc20Addr,
			Erc20Decimals: row.Values[7].(int64),
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

// creditBalance credits a balance to a user.
// The rewardId is the ID of the reward instance.
// It if is negative, it will subtract.
func creditBalance(ctx context.Context, app *common.App, rewardId *types.UUID, user ethcommon.Address, amount *types.Decimal) error {
	balanceId := userBalanceID(rewardId, user)
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}INSERT INTO balances(id, reward_id, address, balance)
	VALUES ($id, $reward_id, $user, $balance)
	ON CONFLICT (id) DO UPDATE SET balance = balances.balance + $balance
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

// transferTokens transfers tokens from one user to another.
func transferTokens(ctx context.Context, app *common.App, rewardID *types.UUID, from, to ethcommon.Address, amount *types.Decimal) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE balances
	SET balance = balance - $amount
	WHERE reward_id = $reward_id AND address = $from;

	{kwil_erc20_meta}INSERT INTO balances(id, reward_id, address, balance)
	VALUES ($to_id, $reward_id, $to, $amount)
	ON CONFLICT (id) DO UPDATE SET balance = balances.balance + $amount;
	`, map[string]any{
		"reward_id": rewardID,
		"from":      from.Bytes(),
		"to":        to.Bytes(),
		"amount":    amount,
		"to_id":     userBalanceID(rewardID, to),
	}, nil)
}

// transferTokensFromUserToNetwork transfers tokens from a user to the network.
func transferTokensFromUserToNetwork(ctx context.Context, app *common.App, rewardID *types.UUID, user ethcommon.Address, amount *types.Decimal) error {
	// we subtract first in case the user does not have enough funds
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE balances
	SET balance = balance - $amount
	WHERE reward_id = $reward_id AND address = $user;

	{kwil_erc20_meta}UPDATE reward_instances
	SET balance = balance + $amount
	WHERE id = $reward_id;
	`, map[string]any{
		"reward_id": rewardID,
		"user":      user.Bytes(),
		"amount":    amount,
	}, nil)
}

// transferTokensFromNetworkToUser transfers tokens from the network to a user.
func transferTokensFromNetworkToUser(ctx context.Context, app *common.App, rewardID *types.UUID, user ethcommon.Address, amount *types.Decimal) error {
	// we subtract first in case the network does not have enough funds
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE reward_instances
	SET balance = balance - $amount
	WHERE id = $reward_id;

	{kwil_erc20_meta}INSERT INTO balances(id, reward_id, address, balance)
	VALUES ($user_id, $reward_id, $user, $amount)
	ON CONFLICT (id) DO UPDATE SET balance = balances.balance + $amount;
	`, map[string]any{
		"reward_id": rewardID,
		"user":      user.Bytes(),
		"amount":    amount,
		"user_id":   userBalanceID(rewardID, user),
	}, nil)
}

// balanceOf gets the balance of a user.
func balanceOf(ctx context.Context, app *common.App, rewardID *types.UUID, user ethcommon.Address) (*types.Decimal, error) {
	var balance *types.Decimal
	err := app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT balance
	FROM balances
	WHERE reward_id = $reward_id AND address = $user
	`, map[string]any{
		"reward_id": rewardID,
		"user":      user.Bytes(),
	}, func(row *common.Row) error {
		if len(row.Values) != 1 {
			return fmt.Errorf("expected 1 value, got %d", len(row.Values))
		}
		balance = row.Values[0].(*types.Decimal)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return balance, nil
}

// getRewardsForEpoch gets all rewards for an epoch.
func getRewardsForEpoch(ctx context.Context, app *common.App, epochID *types.UUID) ([]*EpochReward, error) {
	var rewards []*EpochReward
	err := app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT recipient, amount
	FROM epoch_rewards
	WHERE epoch_id = $epoch_id
	`, map[string]any{
		"epoch_id": epochID,
	}, func(row *common.Row) error {
		if len(row.Values) != 2 {
			return fmt.Errorf("expected 2 values, got %d", len(row.Values))
		}

		recipient, err := bytesToEthAddress(row.Values[0].([]byte))
		if err != nil {
			return err
		}

		rewards = append(rewards, &EpochReward{
			Recipient: recipient,
			Amount:    row.Values[1].(*types.Decimal),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

// getUnconfirmedEpochs gets all unconfirmed epochs.
func getUnconfirmedEpochs(ctx context.Context, app *common.App, instanceID *types.UUID, fn func(*Epoch) error) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	SELECT id, created_at_block, created_at_unix, ended_at, block_hash, reward_root
	FROM epochs
	WHERE instance_id = $instance_id AND confirmed IS FALSE
	ORDER BY ended_at ASC
	`, map[string]any{
		"instance_id": instanceID,
	}, func(r *common.Row) error {
		if len(r.Values) != 6 {
			return fmt.Errorf("expected 6 values, got %d", len(r.Values))
		}

		id := r.Values[0].(*types.UUID)
		createdAtBlock := r.Values[1].(int64)
		createdAtUnix := r.Values[2].(int64)
		endedAt := r.Values[3].(int64)
		blockHash := r.Values[4].([]byte)
		rewardRoot := r.Values[5].([]byte)

		return fn(&Epoch{
			PendingEpoch: PendingEpoch{
				ID:          id,
				StartHeight: createdAtBlock,
				StartTime:   time.Unix(createdAtUnix, 0),
			},
			EndHeight: &endedAt,
			BlockHash: blockHash,
			Root:      rewardRoot,
		})
	})
}

// getVersion gets the version of the meta extension.
func getVersion(ctx context.Context, app *common.App) (version int64, notYetSet bool, err error) {
	count := 0
	err = app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT version
	FROM meta
	`, nil, func(r *common.Row) error {
		if len(r.Values) != 1 {
			return fmt.Errorf("expected 1 value, got %d", len(r.Values))
		}
		count++
		version = r.Values[0].(int64)
		return nil
	})
	switch {
	case errors.Is(err, engine.ErrNamespaceNotFound):
		return 0, true, nil
	case err != nil:
		return 0, false, err
	}

	switch count {
	case 0:
		return 0, true, nil
	case 1:
		return version, false, nil
	default:
		return 0, false, errors.New("expected only one value for version table, got")
	}
}

var currentVersion = int64(1)

// setVersion sets the version of the meta extension.
func setVersionToCurrent(ctx context.Context, app *common.App) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}INSERT INTO meta(version)
	VALUES ($version)
	ON CONFLICT (version) DO UPDATE SET version = $version
	`, map[string]any{
		"version": currentVersion,
	}, nil)
}
