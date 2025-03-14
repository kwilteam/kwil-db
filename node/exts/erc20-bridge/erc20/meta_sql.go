package erc20

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

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
		"created_at_unix":  epoch.StartTime,
		"instance_id":      instanceID,
	}, nil)
}

// finalizeEpoch finalizes an epoch.
// It sets the end height, block hash, and reward root
func finalizeEpoch(ctx context.Context, app *common.App, epochID *types.UUID, endHeight int64, blockHash []byte, root []byte, total *types.Decimal) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE epochs
	SET ended_at = $ended_at,
		block_hash = $block_hash,
		reward_root = $reward_root,
        reward_amount = $reward_amount
	WHERE id = $id
	`, map[string]any{
		"id":            epochID,
		"ended_at":      endHeight,
		"block_hash":    blockHash,
		"reward_root":   root,
		"reward_amount": total,
	}, nil)
}

// confirmEpoch confirms an epoch was received on-chain, also delete all the votes
// associated with the epoch.
func confirmEpoch(ctx context.Context, app *common.App, root []byte) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE epochs
	SET confirmed = true
	WHERE reward_root = $root;

    {kwil_erc20_meta}DELETE FROM epoch_votes where epoch_id=(SELECT id FROM epochs WHERE reward_root = $root);
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

// getStoredRewardInstances gets all stored reward instances. Also returns the
// current epoch(not finalized) that is being used.
func getStoredRewardInstances(ctx context.Context, app *common.App) ([]*rewardExtensionInfo, error) {
	var rewards []*rewardExtensionInfo
	err := app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT r.id, r.chain_id, r.escrow_address, r.distribution_period, r.synced, r.active,
		r.erc20_address, r.erc20_decimals, r.synced_at, r.balance, e.id AS epoch_id,
		e.created_at_block AS epoch_created_at_block, e.created_at_unix AS epoch_created_at_seconds
	FROM reward_instances r
	LEFT JOIN epochs e on r.id = e.instance_id AND e.confirmed IS NOT TRUE AND e.ended_at IS NULL
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
			StartTime:   epochCreatedAtUnix,
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
// If it is negative, it will subtract.
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

// reuseRewardInstance reuse existing synced reward instance, set active status to true,
// and update the distribution period.
// This should be only called when re-use an extension.
func reuseRewardInstance(ctx context.Context, app *common.App, id *types.UUID, distributionPeriod int64) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
    {kwil_erc20_meta}UPDATE reward_instances
    SET distribution_period = $distribution_period, active = true
    WHERE id = $id;
    `, map[string]any{
		"id":                  id,
		"distribution_period": distributionPeriod,
	}, nil)
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
func issueReward(ctx context.Context, app *common.App, instanceId *types.UUID, epochID *types.UUID, user ethcommon.Address, amount *types.Decimal) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}UPDATE reward_instances
	SET balance = balance - $amount
    WHERE id = $instance_id;

	{kwil_erc20_meta}INSERT INTO epoch_rewards(epoch_id, recipient, amount)
	VALUES ($epoch_id, $user, $amount)
	ON CONFLICT (epoch_id, recipient) DO UPDATE SET amount = epoch_rewards.amount + $amount;
	`, map[string]any{
		"instance_id": instanceId,
		"epoch_id":    epochID,
		"user":        user.Bytes(),
		"amount":      amount,
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
func getRewardsForEpoch(ctx context.Context, app *common.App, epochID *types.UUID, fn func(reward *EpochReward) error) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
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

		return fn(&EpochReward{
			Recipient: recipient,
			Amount:    row.Values[1].(*types.Decimal),
		})
	})
}

// previousEpochConfirmed return whether previous exists and confirmed.
func previousEpochConfirmed(ctx context.Context, app *common.App, instanceID *types.UUID, endBlock int64) (exist bool, confirmed bool, err error) {
	err = app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
    {kwil_erc20_meta}SELECT confirmed from epochs
    WHERE instance_id = $instance_id AND ended_at = $end_block
    `, map[string]any{
		"instance_id": instanceID,
		"end_block":   endBlock,
	}, func(r *common.Row) error {
		// might be not necessary
		if exist {
			return fmt.Errorf("internal bug: expected single record")
		}
		exist = true

		if len(r.Values) != 1 {
			return fmt.Errorf("expected 1 values, got %d", len(r.Values))
		}

		confirmed = r.Values[0].(bool)
		return nil
	})

	return exist, confirmed, err
}

func rowToEpoch(r *common.Row) (*Epoch, error) {
	if len(r.Values) != 11 {
		return nil, fmt.Errorf("expected 11 values, got %d", len(r.Values))
	}

	id := r.Values[0].(*types.UUID)
	createdAtBlock := r.Values[1].(int64)
	createdAtUnix := r.Values[2].(int64)

	var rewardRoot []byte
	if r.Values[3] != nil {
		rewardRoot = r.Values[3].([]byte)
	}

	var rewardAmount *types.Decimal
	if r.Values[4] != nil {
		rewardAmount = r.Values[4].(*types.Decimal)
	}

	var endedAt int64
	if r.Values[5] != nil {
		endedAt = r.Values[5].(int64)
	}

	var blockHash []byte
	if r.Values[6] != nil {
		blockHash = r.Values[6].([]byte)
	}

	confirmed := r.Values[7].(bool)

	// NOTE: empty value is [[]]
	// values[8]-values[10] will all be empty if any, from the SQL;
	var voters []ethcommon.Address
	if r.Values[8] != nil {
		rawVoters := r.Values[8].([][]byte)
		// empty value is [[]], cannot use make(), otherwise we'll have a empty `ethcommon.Address`
		for _, rawVoter := range rawVoters {
			if len(rawVoter) == 0 {
				continue
			}
			voter, err := bytesToEthAddress(rawVoter)
			if err != nil {
				return nil, err
			}
			voters = append(voters, voter)
		}
	}

	// NOTE: empty value is [<nil>]
	var voteNonces []int64
	if r.Values[9] != nil {
		rawNonces := r.Values[9].([]*int64)
		for _, rawNonce := range rawNonces {
			if rawNonce != nil {
				// NOTE: this is probably problematic, since we can messup the index
				// If we don't skip, return -1 ?
				voteNonces = append(voteNonces, *rawNonce)
			}
		}
	}

	// NOTE: empty value is [[]]
	var signatures [][]byte
	if r.Values[10] != nil {
		// we skip the empty value, otherwise after conversion, [<nil>] will be returned
		for _, rawSig := range r.Values[10].([][]byte) {
			if len(rawSig) != 0 {
				signatures = append(signatures, rawSig)
			}
		}
	}

	return &Epoch{
		PendingEpoch: PendingEpoch{
			ID:          id,
			StartHeight: createdAtBlock,
			StartTime:   createdAtUnix,
		},
		EndHeight: &endedAt,
		BlockHash: blockHash,
		Root:      rewardRoot,
		Total:     rewardAmount,
		Confirmed: confirmed,
		EpochVoteInfo: EpochVoteInfo{
			Voters:     voters,
			VoteSigs:   signatures,
			VoteNonces: voteNonces,
		},
	}, nil
}

// getActiveEpochs get current active epochs, at most two:
// one collects all new rewards, and one waits to be confirmed.
func getActiveEpochs(ctx context.Context, app *common.App, instanceID *types.UUID, fn func(*Epoch) error) error {
	query := `
    {kwil_erc20_meta}SELECT e.id, e.created_at_block, e.created_at_unix, e.reward_root, e.reward_amount, e.ended_at, e.block_hash, e.confirmed, array_agg(v.voter) as voters, array_agg(v.nonce) as nonces, array_agg(v.signature) as signatures
	FROM epochs AS e
	LEFT JOIN epoch_votes AS v ON v.epoch_id = e.id
	WHERE e.instance_id = $instance_id AND e.confirmed IS NOT true
    GROUP BY e.id, e.created_at_block, e.created_at_unix, e.reward_root, e.reward_amount, e.ended_at, e.block_hash, e.confirmed
    ORDER BY e.created_at_block ASC `
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, query, map[string]any{
		"instance_id": instanceID,
	}, func(r *common.Row) error {
		epoch, err := rowToEpoch(r)
		if err != nil {
			return err
		}
		return fn(epoch)
	})
}

// getEpochs gets epochs.
func getEpochs(ctx context.Context, app *common.App, instanceID *types.UUID, after int64, limit int64, fn func(*Epoch) error) error {
	query := `
    {kwil_erc20_meta}SELECT e.id, e.created_at_block, e.created_at_unix, e.reward_root, e.reward_amount, e.ended_at, e.block_hash, e.confirmed, array_agg(v.voter) as voters, array_agg(v.nonce) as nonces, array_agg(v.signature) as signatures
	FROM epochs AS e
	LEFT JOIN epoch_votes AS v ON v.epoch_id = e.id
	WHERE e.instance_id = $instance_id AND e.created_at_block > $after
    GROUP BY e.id, e.created_at_block, e.created_at_unix, e.reward_root, e.reward_amount, e.ended_at, e.block_hash, e.confirmed
    ORDER BY e.ended_at ASC LIMIT $limit`

	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, query, map[string]any{
		"instance_id": instanceID,
		"after":       after,
		"limit":       limit,
	}, func(r *common.Row) error {
		epoch, err := rowToEpoch(r)
		if err != nil {
			return err
		}
		return fn(epoch)
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

// canVoteEpoch returns a bool indicate whether an epoch can be voted.
func canVoteEpoch(ctx context.Context, app *common.App, epochID *types.UUID) (ok bool, err error) {
	// get epoch that is finalized, but not confirmed.
	err = app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}SELECT confirmed
    FROM epochs WHERE id = $id AND ended_at IS NOT NULL AND confirmed IS NOT true;
    `, map[string]any{
		"id": epochID,
	}, func(row *common.Row) error {
		if len(row.Values) != 1 {
			return fmt.Errorf("expected 1 value, got %d", len(row.Values))
		}

		ok = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return ok, nil
}

// voteEpoch vote an epoch by submitting signature.
// This is idempotent.
func voteEpoch(ctx context.Context, app *common.App, epochID *types.UUID,
	voter ethcommon.Address, nonce int64, signature []byte) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, `
	{kwil_erc20_meta}INSERT into epoch_votes(epoch_id, voter, nonce, signature)
    VALUES ($epoch_id, $voter, $nonce, $signature) ON CONFLICT DO NOTHING;
	`, map[string]any{
		"epoch_id":  epochID,
		"voter":     voter.Bytes(),
		"signature": signature,
		"nonce":     nonce,
	}, nil)
}

// getWalletEpochs returns all confirmed epochs that the given wallet has reward in.
// If pending=true, return all finalized epochs(no necessary confirmed).
func getWalletEpochs(ctx context.Context, app *common.App, instanceID *types.UUID,
	wallet ethcommon.Address, pending bool, fn func(*Epoch) error) error {

	// WE don't need vote info, we just return empty arrays instead of JOIN
	query := `
	{kwil_erc20_meta}SELECT e.id, e.created_at_block, e.created_at_unix, e.reward_root, e.reward_amount, e.ended_at, e.block_hash, e.confirmed, ARRAY[]::BYTEA[] as voters, ARRAY[]::INT8[] as nonces, ARRAY[]::BYTEA[] as signatures
	FROM epoch_rewards AS r
	JOIN epochs AS e ON r.epoch_id = e.id
	WHERE recipient = $wallet AND e.instance_id = $instance_id AND e.ended_at IS NOT NULL` // at least finalized
	if !pending {
		query += ` AND e.confirmed IS true`
	}

	query += ";"
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, query,
		map[string]any{
			"wallet":      wallet.Bytes(),
			"instance_id": instanceID,
		}, func(r *common.Row) error {
			epoch, err := rowToEpoch(r)
			if err != nil {
				return err
			}
			return fn(epoch)
		})
}
