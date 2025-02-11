// package meta implements a meta extension that manages all rewards on a Kwil network.
// It is used to create other extensions with which users can distribute erc20 tokens
// to users.
// It works by exposing an action to the DB owner which allows creation of new extensions
// for specific erc20s. When the action is called, it starts event listeners which sync
// information about the escrow contract, erc20, and multisig from the EVM chain.
// When an extension is in this state, we consider it "pending".
// Once synced, the extension is no longer "pending", but instead ready.
// At this point, users can access the extension's namespace to distribute rewards.
// Internally, the node will start another event listener which is responsible for tracking
// the erc20's Transfer event. When a transfer event is detected, the node will update the
// reward balance of the recipient.
package erc20reward

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/extensions/hooks"
	"github.com/kwilteam/kwil-db/extensions/listeners"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/exts/erc20reward/abigen"
	"github.com/kwilteam/kwil-db/node/exts/erc20reward/reward"
	evmsync "github.com/kwilteam/kwil-db/node/exts/evm-sync"
	"github.com/kwilteam/kwil-db/node/exts/evm-sync/chains"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

const (
	RewardMetaExtensionName = "kwil_erc20_meta"
	uint256Precision        = 78
)

var (
	rewardExtUUIDNamespace = *types.MustParseUUID("b1f140d1-91cf-4bbe-8f78-8f17f6282fc2")
	minEpochPeriod         = time.Minute * 10
	maxEpochPeriod         = time.Hour * 24 * 7 // 1 week
)

// generates a deterministic UUID for the chain and escrow
func uuidForChainAndEscrow(chain string, escrow string) types.UUID {
	return types.NewUUIDV5WithNamespace(rewardExtUUIDNamespace, []byte(chain+escrow))
}

// generates a unique name for the state poller
func statePollerUniqueName(id types.UUID) string {
	return statePollerPrefix + id.String()
}

// idFromStatePollerUniqueName extracts the id from the unique name
func idFromStatePollerUniqueName(name string) (*types.UUID, error) {
	if !strings.HasPrefix(name, statePollerPrefix) {
		return nil, fmt.Errorf("invalid state poller name %s", name)
	}

	return types.ParseUUID(strings.TrimPrefix(name, statePollerPrefix))
}

const (
	statePollerPrefix           = "erc20_state_poll_"
	transferListenerPrefix      = "erc20_transfer_listener_"
	transferEventResolutionName = "erc20_transfer_sync"
	statePollResolutionName     = "erc20_state_poll_sync"
)

// transferListenerUniqueName generates a unique name for the transfer listener
func transferListenerUniqueName(id types.UUID) string {
	return transferListenerPrefix + id.String()
}

// idFromTransferListenerUniqueName extracts the id from the unique name
func idFromTransferListenerUniqueName(name string) (*types.UUID, error) {
	if !strings.HasPrefix(name, transferListenerPrefix) {
		return nil, fmt.Errorf("invalid transfer listener name %s", name)
	}

	return types.ParseUUID(strings.TrimPrefix(name, transferListenerPrefix))
}

// generateEpochID generates a deterministic UUID for an epoch
func generateEpochID(instanceID *types.UUID, startheight int64) *types.UUID {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(startheight))
	id := types.NewUUIDV5WithNamespace(*instanceID, buf)
	return &id
}

/*
This extension is quite complex, because it manages a number of lifetimes and sub extensions.
It is interacted with using "top-level" extensions (e.g. USE erc20 {...} AS alias).
It also interacts with two different event listeners: one for polling state of on-chain contracts
(e.g. the configured erc20 and multisig for an escrow), and one for listening to Transfer events
to reward users with incoming tokens.

Therefore, we break down the extension into "Instances". Instances are defined by a Chain ID and an
escrow address, which is then hashed to create a UUID. This UUID is used to track the extension in the
database. Each instance has its own erc20, multisig, and set of rewards.

Instances have two separate types of states: synchronization and activation.
Synchronization refers to whether we have synced relevant metadata from the EVM chain.
This includes things like the escrow's erc20 address, multisig address, etc.
The second state is activation, which refers to whether the current network is using the extension.
Since extensions can never be fully dropped (as any rewards that are distributed but unclaimed would
effectively be lost), we only deactivate them, and re-activate them when needed.

The most complex part of this extension is the "prepare" method. This method is called when we should
start a new extension, or re-activate a de-activated extension. Therefore, on "prepare", there are 4 states
to consider:
- Extension has never existed: register it in the database and start the state poller
- Extension has existed but is deactivated, and is not synced: start the state poller and inform the DB that it is active
- Extension has existed but is deactivated, and is synced: inform the DB that it is activated and ready, and start the Transfer listener
- Extension has existed and is active: return an error

The other most complex part of this extension is the startup. On startup, we read all existing rewards from the DB,
which may also be in any of the above states. We will store all instance info in memory, and do the following
depending on the state:
- Inactive, Unsynced: do nothing
- Inactive, Synced: do nothing
- Active, Unsynced: start the state poller
- Active, Synced: start the Transfer listener

Upon successful synchronization, the extension is considered "ready".
In other words, ready = synced && activated.
Once an extensiuon is ready, it can be used to distribute rewards to users.
It will also start a listener for Transfer events on the erc20 contract, to update user balances.
*/

func init() {
	/*
		for simplicity, we use a singleton to manage all instances.
		This singleton manages state for all reward instances.
		We can break down everywhere it is referenced into 4 categories:
		1. Extension methods
		2. Resolution extensions (used for resolving synced contract state and events)
		3. End block hooks (used for proposing epochs and resolving ordered events from a listener)
		4. Event listeners (used for listening to events or polling for state on the EVM chain)

		Resolutions and hooks run as part of consensus process.
		Methods _usually_ run as part of consensus, however they can
		run in a read-only context (if marked with VIEW).

		Therefore, we need to account for state being read and written
		concurrently.

		Event listeners run outside of consensus, and thus we have potential
		concurrency issues. All variables provided to event listeners
		are copied; this avoids concurrency issues, as well as ensures that
		the listeners don't cause non-deterministic behavior by modifying
		state.

		I considered making a global singleton instead of defining it here, but I felt
		that it was more clear to track where it was used by defining it here.
	*/

	SINGLETON := &extensionInfo{
		instances: make(map[types.UUID]*rewardExtensionInfo),
	}

	evmsync.RegisterEventResolution(transferEventResolutionName, func(ctx context.Context, app *common.App, block *common.BlockContext, uniqueName string, logs []ethtypes.Log) error {
		id, err := idFromTransferListenerUniqueName(uniqueName)
		if err != nil {
			return err
		}

		for _, log := range logs {
			err := applyTransferLog(ctx, app, id, log)
			if err != nil {
				return err
			}
		}

		return nil
	})

	evmsync.RegisterStatePollResolution(statePollResolutionName, func(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext, uniqueName string, decodedData []byte) error {
		id, err := idFromStatePollerUniqueName(uniqueName)
		if err != nil {
			return err
		}

		SINGLETON.mu.Lock()
		defer SINGLETON.mu.Unlock()

		info, ok := SINGLETON.instances[*id]
		if !ok {
			return fmt.Errorf("reward extension with id %s not found", id)
		}
		if info.synced {
			// signals a serious internal bug
			return fmt.Errorf("duplicate sync resolution for extension with id %s", id)
		}

		var data syncedRewardData
		err = data.UnmarshalBinary(decodedData)
		if err != nil {
			return fmt.Errorf("failed to unmarshal synced reward data: %v", err)
		}

		err = setRewardSynced(ctx, app, id, block.Height, &data)
		if err != nil {
			return err
		}

		info.synced = true
		info.syncedAt = block.Height
		info.syncedRewardData = data

		err = evmsync.StatePoller.UnregisterPoll(uniqueName)
		if err != nil {
			return err
		}

		// if active, we should start the transfer listener
		// Otherwise, we will just wait until it is activated
		if info.active {
			return info.startTransferListener(ctx, app)
		}

		return nil
	})

	err := precompiles.RegisterInitializer(RewardMetaExtensionName,
		func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (precompiles.Precompile, error) {
			return precompiles.Precompile{
				Cache: SINGLETON,
				OnUse: func(ctx *common.EngineContext, app *common.App) error {
					return createSchema(ctx.TxContext.Ctx, app)
				},
				OnStart: func(ctx context.Context, app *common.App) error {
					// if the schema exists, we should read all existing reward instances
					instances, err := getStoredRewardInstances(ctx, app)
					switch {
					case err == nil:
						// do nothing
					case errors.Is(err, engine.ErrNamespaceNotFound):
						// if the schema doesnt exist, then we just return
						// since genesis has not been run yet
						return nil
					default:
						return err
					}

					SINGLETON.mu.Lock()
					defer SINGLETON.mu.Unlock()

					for _, instance := range instances {
						// if instance is active, we should start one of its
						// two listeners. If it is synced, we should start the
						// transfer listener. Otherwise, we should start the state poller
						if instance.active {
							if instance.synced {
								err = instance.startTransferListener(ctx, app)
								if err != nil {
									return err
								}
							} else {
								err = instance.startStatePoller()
								if err != nil {
									return err
								}
							}
						}

						SINGLETON.instances[*instance.ID] = instance
					}

					return nil
				},
				Methods: []precompiles.Method{
					{
						// prepare begins the sync process for a new reward extension.
						Name: "prepare",
						Parameters: []precompiles.PrecompileValue{
							{Name: "chain", Type: types.TextType},
							{Name: "escrow", Type: types.TextType},
							{Name: "period", Type: types.TextType},
						},
						Returns: &precompiles.MethodReturn{
							Fields: []precompiles.PrecompileValue{
								{Name: "id", Type: types.UUIDType},
							},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) (err error) {
							chain := inputs[0].(string)
							escrow := inputs[1].(string)
							period := inputs[2].(string)

							if !ethcommon.IsHexAddress(escrow) {
								return fmt.Errorf("escrow address %s is not a valid ethereum address", escrow)
							}
							escrowAddress := ethcommon.HexToAddress(escrow)

							id := uuidForChainAndEscrow(chain, escrow)

							dur, err := time.ParseDuration(period) // ensure period is a valid time
							if err != nil {
								return err
							}

							if dur < minEpochPeriod || dur > maxEpochPeriod {
								return fmt.Errorf("epoch period %s is not within the range [%s, %s]", dur, minEpochPeriod, maxEpochPeriod)
							}

							trunced := dur.Truncate(time.Second) // truncate to seconds
							if trunced != dur {
								return fmt.Errorf("epoch period %s is not a whole number of seconds", dur)
							}

							chainConst := chains.Chain(chain) // ensure chain exists
							err = chainConst.Valid()
							if err != nil {
								return err
							}

							chainInfo, ok := chains.GetChainInfo(chainConst)
							if !ok {
								return fmt.Errorf("chain with name %s not found", chain)
							}

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							info, ok := SINGLETON.instances[id]
							// if the instance already exists, it can be in two states:
							// 1. active: we should return an error
							// 2. inactive
							// If inactive, we should check if it is synced. If it is, we should
							// start the transfer listener. Otherwise, we should get it synced by
							// starting the state poller.
							if ok {
								if info.active {
									return fmt.Errorf(`reward extension with chain "%s" and escrow "%s" is already active`, chain, escrow)
								}
								if info.synced {
									// if it is already synced, we should just make sure to start listening
									// to transfer events and activate it

									err = setActiveStatus(ctx.TxContext.Ctx, app, &id, true)
									if err != nil {
										return err
									}
									info.active = true

									err = info.startTransferListener(ctx.TxContext.Ctx, app)
									if err != nil {
										return err
									}

									return resultFn([]any{id})
								}
								// do nothing, we will proceed below to start the state poller
							} else {
								firstEpoch := newPendingEpoch(&id, ctx.TxContext.BlockContext)
								// if not synced, register the new reward extension
								info = &rewardExtensionInfo{
									userProvidedData: userProvidedData{
										ID:                 &id,
										ChainInfo:          &chainInfo,
										EscrowAddress:      escrowAddress,
										DistributionPeriod: int64(dur.Seconds()),
									},
									currentEpoch: firstEpoch,
								}

								err = createNewRewardInstance(ctx.TxContext.Ctx, app, &info.userProvidedData)
								if err != nil {
									return err
								}

								// create the first epoch
								err = createEpoch(ctx.TxContext.Ctx, app, firstEpoch, &id)
								if err != nil {
									return err
								}
							}

							err = info.startStatePoller()
							if err != nil {
								return err
							}

							// we wait until here to add it in case there is an error
							// in RegisterPoll. This only matters if it is new, otherwise
							// we are just setting the same info in the map again
							SINGLETON.instances[id] = info

							return resultFn([]any{id})
						},
					},
					{
						// disable disables a reward extension.
						Name: "disable",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							info, ok := SINGLETON.instances[*id]
							if !ok {
								return fmt.Errorf("reward extension with id %s not found", id)
							}

							if !info.active {
								return fmt.Errorf("reward extension with id %s is already disabled", id)
							}

							err := setActiveStatus(ctx.TxContext.Ctx, app, id, false)
							if err != nil {
								return err
							}

							err = info.stopAllListeners()
							if err != nil {
								return err
							}

							info.active = false

							return nil
						},
					},
					{
						// info returns information about a reward extension.
						Name: "info",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
						},
						Returns: &precompiles.MethodReturn{
							Fields: []precompiles.PrecompileValue{
								{Name: "chain", Type: types.TextType},
								{Name: "escrow", Type: types.TextType},
								{Name: "epoch_period", Type: types.IntType},
								{Name: "erc20", Type: types.TextType, Nullable: true},
								{Name: "decimals", Type: types.IntType, Nullable: true},
								{Name: "balance", Type: types.TextType}, // total unspent balance
								{Name: "synced", Type: types.BoolType},
								{Name: "synced_at", Type: types.IntType, Nullable: true},
								{Name: "enabled", Type: types.BoolType},
							},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)

							SINGLETON.mu.RLock()
							defer SINGLETON.mu.RUnlock()

							info, ok := SINGLETON.instances[*id]
							if !ok {
								return fmt.Errorf("reward extension with id %s not found", id)
							}

							erc20 := info.syncedRewardData.Erc20Address.Hex()
							decimals := info.syncedRewardData.Erc20Decimals

							return resultFn([]any{
								info.userProvidedData.ChainInfo,
								info.userProvidedData.EscrowAddress.Hex(),
								info.userProvidedData.DistributionPeriod,
								erc20,
								decimals,
								info.ownedBalance.String(),
								info.synced,
								info.syncedAt,
								info.active,
							})
						},
					},
					{
						// id returns the ID of a reward extension.
						Name: "id",
						Parameters: []precompiles.PrecompileValue{
							{Name: "chain", Type: types.TextType},
							{Name: "escrow", Type: types.TextType},
						},
						Returns: &precompiles.MethodReturn{
							Fields: []precompiles.PrecompileValue{
								{Name: "id", Type: types.UUIDType},
							},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							chain := inputs[0].(string)
							escrow := inputs[1].(string)

							id := uuidForChainAndEscrow(chain, escrow)

							return resultFn([]any{id})
						},
					},
					{
						// list returns a list of all reward extensions.
						Name: "list",
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields: []precompiles.PrecompileValue{
								{Name: "id", Type: types.UUIDType},
								{Name: "chain", Type: types.TextType},
								{Name: "escrow", Type: types.TextType},
							},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							SINGLETON.mu.RLock()
							defer SINGLETON.mu.RUnlock()

							return SINGLETON.ForEachInstance(func(id *types.UUID, info *rewardExtensionInfo) error {
								return resultFn([]any{id, info.userProvidedData.ChainInfo, info.userProvidedData.EscrowAddress.Hex()})
							})
						},
					},
					{
						// issue issues a reward to a user.
						Name: "issue",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
							{Name: "user", Type: types.TextType},
							{Name: "amount", Type: types.TextType},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)
							user := inputs[1].(string)
							amount := inputs[2].(string)

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							info, err := SINGLETON.getUsableInstance(id)
							if err != nil {
								return err
							}

							// users will pass rewards with the proper number of decimals.
							// E.g. if the erc20 has 18 decimals, they will pass 18 decimals
							rawAmount, err := parseAmountFromUser(amount, uint16(info.syncedRewardData.Erc20Decimals))
							if err != nil {
								return err
							}

							if rawAmount.IsNegative() {
								return fmt.Errorf("amount cannot be negative")
							}

							newBal, err := types.DecimalSub(info.ownedBalance, rawAmount)
							if err != nil {
								return err
							}

							if newBal.IsNegative() {
								return fmt.Errorf("network does not enough balance to issue %s to %s", amount, user)
							}

							addr, err := ethAddressFromHex(user)
							if err != nil {
								return err
							}

							err = issueReward(ctx.TxContext.Ctx, app, info.currentEpoch.ID, addr, rawAmount)
							if err != nil {
								return err
							}

							info.ownedBalance = newBal

							return nil
						},
					},
					{
						// transfer transfers tokens from the caller to another address.
						Name: "transfer",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
							{Name: "to", Type: types.TextType},
							{Name: "amount", Type: types.TextType},
						},
						// anybody can call this as long as they have the tokens.
						// There is no security risk if somebody calls this directly
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)
							to := inputs[1].(string)
							amount := inputs[2].(string)

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							info, err := SINGLETON.getUsableInstance(id)
							if err != nil {
								return err
							}

							// users will pass rewards with the proper number of decimals.
							// E.g. if the erc20 has 18 decimals, they will pass 18 decimals
							rawAmount, err := parseAmountFromUser(amount, uint16(info.syncedRewardData.Erc20Decimals))
							if err != nil {
								return err
							}

							if rawAmount.IsNegative() {
								return fmt.Errorf("amount cannot be negative")
							}

							from, err := ethAddressFromHex(ctx.TxContext.Caller)
							if err != nil {
								return err
							}

							toAddr, err := ethAddressFromHex(to)
							if err != nil {
								return err
							}

							return transferTokens(ctx.TxContext.Ctx, app, id, from, toAddr, rawAmount)
						},
					},
					{
						// locks takes tokens from a user's balance and gives them to the network.
						// The network can then distribute these tokens to other users, either via
						// unlock or issue.
						Name: "lock",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
							{Name: "amount", Type: types.TextType},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)
							amount := inputs[1].(string)

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							return SINGLETON.lockTokens(ctx.TxContext.Ctx, app, id, ctx.TxContext.Caller, amount)
						},
					},
					{
						// lock_admin is a privileged version of lock that allows the network to lock
						// tokens on behalf of a user.
						Name: "lock_admin",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
							{Name: "user", Type: types.TextType},
							{Name: "amount", Type: types.TextType},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)
							user := inputs[1].(string)
							amount := inputs[2].(string)

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							return SINGLETON.lockTokens(ctx.TxContext.Ctx, app, id, user, amount)
						},
					},
					{
						// unlock returns tokens to a user's balance that were previously locked.
						// It returns the tokens so that the user can spend them.
						// It can only be called by the network when it wishes to return tokens to a user.
						Name: "unlock",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
							{Name: "user", Type: types.TextType},
							{Name: "amount", Type: types.TextType},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)
							user := inputs[1].(string)
							amount := inputs[2].(string)

							SINGLETON.mu.Lock()
							defer SINGLETON.mu.Unlock()

							info, err := SINGLETON.getUsableInstance(id)
							if err != nil {
								return err
							}

							// users will pass rewards with the proper number of decimals.
							// E.g. if the erc20 has 18 decimals, they will pass 18 decimals
							rawAmount, err := parseAmountFromUser(amount, uint16(info.syncedRewardData.Erc20Decimals))
							if err != nil {
								return err
							}

							if rawAmount.IsNegative() {
								return fmt.Errorf("amount cannot be negative")
							}

							addr, err := ethAddressFromHex(user)
							if err != nil {
								return err
							}

							left, err := types.DecimalSub(info.ownedBalance, rawAmount)
							if err != nil {
								return err
							}

							if left.IsNegative() {
								return fmt.Errorf("network does not have enough balance to unlock %s for %s", amount, user)
							}

							err = transferTokensFromNetworkToUser(ctx.TxContext.Ctx, app, id, addr, rawAmount)
							if err != nil {
								return err
							}

							info.ownedBalance = left
							return nil
						},
					},
					{
						// balance returns the balance of a user.
						Name: "balance",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
							{Name: "user", Type: types.TextType},
						},
						Returns: &precompiles.MethodReturn{
							Fields: []precompiles.PrecompileValue{
								{Name: "balance", Type: types.TextType},
							},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)
							user := inputs[1].(string)

							addr, err := ethAddressFromHex(user)
							if err != nil {
								return err
							}

							bal, err := balanceOf(ctx.TxContext.Ctx, app, id, addr)
							if err != nil {
								return err
							}

							return resultFn([]any{bal.String()})
						},
					},
					{
						// lists epochs that have not been confirmed yet, but have been ended.
						// It lists them in ascending order (e.g. oldest first).
						Name: "list_unconfirmed_epochs",
						Parameters: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields: []precompiles.PrecompileValue{
								{Name: "epoch_id", Type: types.UUIDType},
								{Name: "start_height", Type: types.IntType},
								{Name: "start_timestamp", Type: types.IntType},
								{Name: "end_height", Type: types.IntType},
								{Name: "reward_root", Type: types.ByteaType},
								{Name: "end_block_hash", Type: types.ByteaType},
							},
						},
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							id := inputs[0].(*types.UUID)

							return getUnconfirmedEpochs(ctx.TxContext.Ctx, app, id, func(e *Epoch) error {
								return resultFn([]any{e.ID, e.StartHeight, e.StartTime.Unix(), *e.EndHeight, e.Root, e.BlockHash})
							})
						},
					},
					// {
					// 	// Supposed to be called by Signer service
					// 	// Returns epoch rewards after(non-include) after_height, in ASC order.
					// 	Name: "list_epochs",
					// 	Parameters: []precompiles.PrecompileValue{
					// 		{Name: "id", Type: types.UUIDType},
					// 		{Name: "after_height", Type: types.IntType},
					// 		{Name: "limit", Type: types.IntType},
					// 	},
					// 	Returns: &precompiles.MethodReturn{
					// 		IsTable: true,
					// 		Fields:  (&Epoch{}).UnpackTypes(), // TODO: I might need to update this depending on what happens with decimal types
					// 	},
					// 	AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
					// 	Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
					// 		panic("finish me")
					// 	},
					// },
					// {
					// 	// Supposed to be called by the SignerService, to verify the reward root.
					// 	// Could be merged into 'list_epochs'
					// 	// Returns pending rewards from(include) start_height to(include) end_height, in ASC order.
					// 	// NOTE: Rewards of same address will be aggregated.
					// 	Name: "search_rewards",
					// 	Parameters: []precompiles.PrecompileValue{
					// 		{Name: "id", Type: types.UUIDType},
					// 		{Name: "start_height", Type: types.IntType},
					// 		{Name: "end_height", Type: types.IntType},
					// 	},
					// 	Returns: &precompiles.MethodReturn{
					// 		IsTable: true,
					// 		Fields:  (&Reward{}).UnpackTypes(), // TODO: I might need to update this depending on what happens with decimal types
					// 	},
					// 	AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
					// 	Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
					// 		panic("finish me")
					// 	},
					// },
					// {
					// 	// Supposed to be called by Kwil network in an end block hook.
					// 	Name:            "propose_epoch",
					// 	AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
					// 	Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
					// 		if !calledByExtension(ctx) {
					// 			return errors.New("propose_epoch can only be called by the Kwil network")
					// 		}

					// 		panic("finish me")
					// 	},
					// },
					// {
					// 	// Supposed to be called by a multisig signer
					// 	Name: "vote_epoch",
					// 	Parameters: []precompiles.PrecompileValue{
					// 		{Name: "id", Type: types.UUIDType},
					// 		{Name: "sign_hash", Type: types.ByteaType},
					// 		{Name: "signatures", Type: types.ByteaType},
					// 	},
					// 	AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
					// 	Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
					// 		panic("finish me")
					// 	},
					// },
					// {
					// 	// Lists all epochs that have received enough votes
					// 	Name: "list_finalized",
					// 	Parameters: []precompiles.PrecompileValue{
					// 		{Name: "id", Type: types.UUIDType},
					// 		{Name: "after_height", Type: types.IntType},
					// 		{Name: "limit", Type: types.IntType},
					// 	},
					// 	Returns: &precompiles.MethodReturn{
					// 		IsTable: true,
					// 		Fields:  (&FinalizedReward{}).UnpackTypes(), // TODO: I might need to update this depending on what happens with decimal types
					// 	},
					// 	AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
					// 	Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
					// 		panic("finish me")
					// 	},
					// },
					// {
					// 	// Called by app / user
					// 	Name: "claim_info",
					// 	Parameters: []precompiles.PrecompileValue{
					// 		{Name: "id", Type: types.UUIDType},
					// 		{Name: "sign_hash", Type: types.ByteaType, Nullable: false},
					// 		{Name: "wallet_address", Type: types.TextType, Nullable: false},
					// 	},
					// 	Returns: &precompiles.MethodReturn{
					// 		Fields: []precompiles.PrecompileValue{
					// 			{Name: "amount", Type: types.TextType},
					// 			{Name: "block_hash", Type: types.TextType},
					// 			{Name: "root", Type: types.TextType},
					// 			{Name: "proofs", Type: types.TextArrayType},
					// 		},
					// 	},
					// 	AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
					// 	Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
					// 		panic("finish me")
					// 	},
					// },
				},
			}, nil
		})
	if err != nil {
		panic(err)
	}

	// we will create the schema at genesis
	err = hooks.RegisterGenesisHook(RewardMetaExtensionName+"_genesis", func(ctx context.Context, app *common.App, chain *common.ChainContext) error {
		version, notYetSet, err := getVersion(ctx, app)
		if err != nil {
			return err
		}
		if notYetSet {
			err = genesisExec(ctx, app)
			if err != nil {
				return err
			}

			err = setVersionToCurrent(ctx, app)
			if err != nil {
				return err
			}
		} else {
			// in the future, we will handle version upgrades here
			if version != currentVersion {
				return fmt.Errorf("reward extension version mismatch: expected %d, got %d", currentVersion, version)
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	// the end block hook will be used to propose epochs
	err = hooks.RegisterEndBlockHook(RewardMetaExtensionName+"_end_block", func(ctx context.Context, app *common.App, block *common.BlockContext) error {
		SINGLETON.mu.Lock()
		defer SINGLETON.mu.Unlock()

		return SINGLETON.ForEachInstance(func(id *types.UUID, info *rewardExtensionInfo) error {
			// if the block is greater than or equal to the start time + distribution period,
			// we should propose a new epoch. Otherwise, we should do nothing.
			if block.Timestamp < info.currentEpoch.StartTime.Add(time.Duration(info.userProvidedData.DistributionPeriod)*time.Second).Unix() {
				return nil
			}

			rewards, err := getRewardsForEpoch(ctx, app, info.currentEpoch.ID)
			if err != nil {
				return err
			}

			users := make([]string, len(rewards))
			amounts := make([]*big.Int, len(rewards))

			for i, reward := range rewards {
				users[i] = reward.Recipient.Hex()
				amounts[i] = reward.Amount.BigInt()
			}

			_, root, err := reward.GenRewardMerkleTree(users, amounts, info.EscrowAddress.Hex(), block.Hash)
			if err != nil {
				return err
			}

			err = finalizeEpoch(ctx, app, info.currentEpoch.ID, block.Height, block.Hash[:], root)
			if err != nil {
				return err
			}

			// create a new epoch
			newEpoch := newPendingEpoch(id, block)
			err = createEpoch(ctx, app, newEpoch, id)
			if err != nil {
				return err
			}

			info.currentEpoch = newEpoch

			return nil
		})
	})
	if err != nil {
		panic(err)
	}
}

func genesisExec(ctx context.Context, app *common.App) error {
	// we will create the schema at genesis
	err := app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, "USE kwil_erc20_meta AS kwil_erc20_meta", nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func callPrepare(ctx *common.EngineContext, app *common.App, chain string, escrow string, period string) (*types.UUID, error) {
	var id *types.UUID
	count := 0
	_, err := app.Engine.Call(ctx, app.DB, RewardMetaExtensionName, "prepare", []any{chain, escrow, period}, func(r *common.Row) error {
		if count > 0 {
			return fmt.Errorf("internal bug: expected only one result on prepare erc20")
		}
		var ok bool
		id, ok = r.Values[0].(*types.UUID)
		if !ok {
			return fmt.Errorf("internal bug: expected UUID")
		}

		count++
		return nil
	})
	return id, err
}

func callDisable(ctx *common.EngineContext, app *common.App, id *types.UUID) error {
	_, err := app.Engine.Call(ctx, app.DB, RewardMetaExtensionName, "disable", []any{id}, nil)
	return err
}

// lockTokens locks tokens from a user's balance and gives them to the network.
func (e *extensionInfo) lockTokens(ctx context.Context, app *common.App, id *types.UUID, from string, amount string) error {
	fromAddr, err := ethAddressFromHex(from)
	if err != nil {
		return err
	}

	info, err := e.getUsableInstance(id)
	if err != nil {
		return err
	}

	rawAmount, err := parseAmountFromUser(amount, uint16(info.syncedRewardData.Erc20Decimals))
	if err != nil {
		return err
	}

	if rawAmount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}

	err = transferTokensFromUserToNetwork(ctx, app, id, fromAddr, rawAmount)
	if err != nil {
		return err
	}

	info.ownedBalance, err = info.ownedBalance.Add(info.ownedBalance, rawAmount)
	if err != nil {
		return err
	}

	return nil
}

// getUsableInstance gets an instance and ensures it is active and synced.
// It is not thread safe and should be called within a lock.
func (e *extensionInfo) getUsableInstance(id *types.UUID) (*rewardExtensionInfo, error) {
	info, ok := e.instances[*id]
	if !ok {
		return nil, fmt.Errorf("reward extension with id %s not found", id)
	}

	if !info.active {
		return nil, fmt.Errorf("reward extension with id %s is not active", id)
	}

	if !info.synced {
		return nil, fmt.Errorf("reward extension with id %s is not synced", id)
	}

	return info, nil
}

func ethAddressFromHex(s string) (ethcommon.Address, error) {
	if !ethcommon.IsHexAddress(s) {
		return ethcommon.Address{}, fmt.Errorf("invalid ethereum address: %s", s)
	}
	return ethcommon.HexToAddress(s), nil
}

// newPendingEpoch creates a new epoch.
func newPendingEpoch(rewardID *types.UUID, block *common.BlockContext) *PendingEpoch {
	return &PendingEpoch{
		ID:          generateEpochID(rewardID, block.Height),
		StartHeight: block.Height,
		StartTime:   time.Unix(block.Timestamp, 0),
	}
}

// PendingEpoch is an epoch that has been started but not yet finalized.
type PendingEpoch struct {
	ID          *types.UUID
	StartHeight int64
	StartTime   time.Time
}

// EpochReward is a reward given to a user within an epoch
type EpochReward struct {
	Recipient ethcommon.Address
	Amount    *types.Decimal // numeric(78, 0)
}

func (p *PendingEpoch) copy() *PendingEpoch {
	id := *p.ID
	return &PendingEpoch{
		ID:          &id,
		StartHeight: p.StartHeight,
	}
}

// Epoch is a period in which rewards are distributed.
type Epoch struct {
	PendingEpoch
	EndHeight *int64 // nil if not finalized
	BlockHash []byte // hash of the block that finalized the epoch, nil if not finalized
	Root      []byte // merkle root of all rewards, nil if not finalized
}

type extensionInfo struct {
	// mu protects all fields in the struct
	mu sync.RWMutex
	// instances tracks all child reward extensions
	instances map[types.UUID]*rewardExtensionInfo
}

// Copy implements the precompiles.Cache interface.
func (e *extensionInfo) Copy() precompiles.Cache {
	e.mu.RLock()
	defer e.mu.RUnlock()

	instances := make(map[types.UUID]*rewardExtensionInfo)
	for k, v := range e.instances {
		instances[k] = v.copy()
	}

	return &extensionInfo{
		instances: instances,
	}
}

func (e *extensionInfo) Apply(v precompiles.Cache) {
	e.mu.Lock()
	defer e.mu.Unlock()

	info := v.(*extensionInfo)
	e.instances = info.instances
}

// ForEachInstance deterministically iterates over all instances of the extension.
func (e *extensionInfo) ForEachInstance(fn func(id *types.UUID, info *rewardExtensionInfo) error) error {
	orderableMap := make(map[string]*rewardExtensionInfo)
	for k, v := range e.instances {
		orderableMap[k.String()] = v
	}

	for _, kv := range order.OrderMap(orderableMap) {
		err := fn(kv.Value.userProvidedData.ID, kv.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// userProvidedData holds information about a reward that is known as soon
// as the `create` action is called.
type userProvidedData struct {
	ID                 *types.UUID       // auto-generated
	ChainInfo          *chains.ChainInfo // chain ID of the EVM chain
	EscrowAddress      ethcommon.Address // address of the escrow contract
	DistributionPeriod int64             // period (in seconds) between reward distributions
}

func (u *userProvidedData) copy() *userProvidedData {
	id := *u.ID
	cInfo := *u.ChainInfo
	return &userProvidedData{
		ID:                 &id,
		ChainInfo:          &cInfo,
		EscrowAddress:      u.EscrowAddress,
		DistributionPeriod: u.DistributionPeriod,
	}
}

// syncedRewardData holds information about a reward that is synced from
// on chain.
type syncedRewardData struct {
	Erc20Address  ethcommon.Address // address of the erc20 contract
	Erc20Decimals int64             // decimals of the erc20 contract
}

func (s *syncedRewardData) copy() *syncedRewardData {
	return &syncedRewardData{
		Erc20Address:  s.Erc20Address,
		Erc20Decimals: s.Erc20Decimals,
	}
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (s *syncedRewardData) MarshalBinary() ([]byte, error) {
	// Allocate 28 bytes: 20 for the address, 8 for the int64.
	b := make([]byte, 28)
	// Copy the address bytes into the first 20 bytes.
	copy(b[:20], s.Erc20Address[:])
	// Encode Erc20Decimals into the next 8 bytes using BigEndian.
	binary.BigEndian.PutUint64(b[20:], uint64(s.Erc20Decimals))
	return b, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (s *syncedRewardData) UnmarshalBinary(data []byte) error {
	// Check that the data is exactly 28 bytes.
	if len(data) != 28 {
		return fmt.Errorf("invalid data length: expected 28, got %d", len(data))
	}
	// Copy the first 20 bytes into the address.
	copy(s.Erc20Address[:], data[:20])
	// Decode the last 8 bytes into Erc20Decimals.
	s.Erc20Decimals = int64(binary.BigEndian.Uint64(data[20:28]))
	return nil
}

// rewardExtensionInfo holds information about a reward extension
type rewardExtensionInfo struct {
	userProvidedData
	syncedRewardData
	synced       bool
	syncedAt     int64
	active       bool
	ownedBalance *types.Decimal // balance owned by DB that can be distributed
	currentEpoch *PendingEpoch  // current epoch being proposed
}

func (r *rewardExtensionInfo) copy() *rewardExtensionInfo {
	decCopy := types.MustParseDecimalExplicit(r.ownedBalance.String(), 78, 0)
	return &rewardExtensionInfo{
		userProvidedData: *r.userProvidedData.copy(),
		syncedRewardData: *r.syncedRewardData.copy(),
		synced:           r.synced,
		syncedAt:         r.syncedAt,
		active:           r.active,
		ownedBalance:     decCopy,
		currentEpoch:     r.currentEpoch.copy(),
	}
}

// startStatePoller starts a state poller for the reward extension.
func (r *rewardExtensionInfo) startStatePoller() error {
	synced := r.synced // copy to avoid race conditions
	escrow := r.EscrowAddress
	id := *r.ID
	chainName := r.ChainInfo.Name

	return evmsync.StatePoller.RegisterPoll(evmsync.PollConfig{
		Chain: chainName,
		PollFunc: func(ctx context.Context, service *common.Service, eventstore listeners.EventKV, broadcast func(context.Context, []byte) error, client *ethclient.Client) {
			// It is _very_ important that we do not change the state of the struct here.
			// This function runs external to consensus, so we must not change the state of the struct.
			if synced {
				return
			}

			data, err := getSyncedRewardData(ctx, client, escrow)
			if err != nil {
				logger := service.Logger.New(statePollerUniqueName(id))
				logger.Errorf("failed to get synced reward data: %v", err)
				return
			}

			bts, err := data.MarshalBinary()
			if err != nil {
				panic(err) // internal logic bug in this package
			}

			err = broadcast(ctx, bts)
			if err != nil {
				logger := service.Logger.New(statePollerUniqueName(id))
				logger.Errorf("failed to get broadcast reward data to network: %v", err)
				return
			}

			synced = true
			// we dont update *rewardExtensionInfo here because we are outside of the consensus process.
			// It will be updated in the resolveFunc
		},
		UniqueName: statePollerUniqueName(*r.ID),
	})
}

// startTransferListener starts an event listener that listens for Transfer events
func (r *rewardExtensionInfo) startTransferListener(ctx context.Context, app *common.App) error {
	// I'm not sure if copies are needed because the values should never be modified,
	// but just in case, I copy them to be used in GetLogs, which runs outside of consensus
	escrowCopy := r.EscrowAddress
	erc20Copy := r.Erc20Address

	// we now register synchronization of the Transfer event
	return evmsync.EventSyncer.RegisterNewListener(ctx, app.DB, app.Engine, evmsync.EVMEventListenerConfig{
		UniqueName: transferListenerUniqueName(*r.ID),
		Chain:      r.ChainInfo.Name,
		GetLogs: func(ctx context.Context, client *ethclient.Client, startBlock, endBlock uint64, logger log.Logger) ([]ethtypes.Log, error) {
			filt, err := abigen.NewErc20Filterer(erc20Copy, client)
			if err != nil {
				return nil, fmt.Errorf("failed to bind to ERC20 filterer: %w", err)
			}

			iter, err := filt.FilterTransfer(&bind.FilterOpts{
				Start:   startBlock,
				End:     &endBlock,
				Context: ctx,
			}, nil, []ethcommon.Address{escrowCopy})
			if err != nil {
				return nil, fmt.Errorf("failed to get transfer logs: %w", err)
			}
			defer iter.Close()

			var logs []ethtypes.Log
			for iter.Next() {
				logs = append(logs, iter.Event.Raw)
			}

			return logs, iter.Error()
		},
		Resolve: transferEventResolutionName,
	})
}

// stopAllListeners stops all event listeners for the reward extension.
// If it is synced, this means it must have an active Transfer listener.
// If it is not synced, it must have an active state poller.
func (r *rewardExtensionInfo) stopAllListeners() error {
	if r.synced {
		return evmsync.EventSyncer.UnregisterListener(transferListenerUniqueName(*r.ID))
	}
	return evmsync.StatePoller.UnregisterPoll(statePollerUniqueName(*r.ID))
}

// nilEthFilterer is a dummy filterer that does nothing.
// Abigen requires a filter to be passed in order to parse event info from logs,
// however the client itself is never actually used.
type nilEthFilterer struct{}

func (nilEthFilterer) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ethtypes.Log, error) {
	return nil, fmt.Errorf("filter logs was not expected to be called")
}

func (nilEthFilterer) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- ethtypes.Log) (ethereum.Subscription, error) {
	return nil, fmt.Errorf("subscribe filter logs was not expected to be called")
}

// applyTransferLog applies a Transfer log to the reward extension.
func applyTransferLog(ctx context.Context, app *common.App, id *types.UUID, log ethtypes.Log) error {
	data, err := erc20LogParser.ParseTransfer(log)
	if err != nil {
		return fmt.Errorf("failed to parse Transfer event: %w", err)
	}

	val, err := erc20ValueFromBigInt(data.Value)
	if err != nil {
		return fmt.Errorf("failed to convert big.Int to decimal.Decimal: %w", err)
	}

	return creditBalance(ctx, app, id, data.From, val)
}

// erc20ValueFromBigInt converts a big.Int to a decimal.Decimal(78,0)
func erc20ValueFromBigInt(b *big.Int) (*types.Decimal, error) {
	dec, err := types.NewDecimalFromBigInt(b, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to convert big.Int to decimal.Decimal: %w", err)
	}
	err = dec.SetPrecisionAndScale(78, 0)
	return dec, err
}

var (
	// erc20LogParser is a pre-bound ERC20 filterer for parsing Transfer events.
	// It should only be used for parsing, since it will not work for reading
	// logs from the EVM chain.
	erc20LogParser = func() ierc20LogParser {
		filt, err := abigen.NewErc20Filterer(ethcommon.Address{}, nilEthFilterer{})
		if err != nil {
			panic(fmt.Sprintf("failed to bind to ERC20 filterer: %v", err))
		}

		return filt
	}()
)

// ierc20LogParser is an interface for parsing ERC20 logs.
// It is only defined to make it clear that the implementation
// should not use other methods of the abigen erc20 type.
type ierc20LogParser interface {
	ParseApproval(log ethtypes.Log) (*abigen.Erc20Approval, error)
	ParseTransfer(log ethtypes.Log) (*abigen.Erc20Transfer, error)
}

// getSyncedRewardData reads on-chain data from both the RewardDistributor and the Gnosis Safe
// it references, returning them in a syncedRewardData struct.
// It does not get the tokens owned by escrow; it will later sync those from erc20 logs
func getSyncedRewardData(
	ctx context.Context,
	client *ethclient.Client,
	distributorAddr ethcommon.Address,
) (*syncedRewardData, error) {

	// 1) Instantiate a binding to RewardDistributor at distributorAddr.
	distributor, err := abigen.NewRewardDistributor(distributorAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to RewardDistributor: %w", err)
	}

	// 2) Read the rewardToken address from the RewardDistributor
	rewardTokenAddr, err := distributor.RewardToken(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to get rewardToken from RewardDistributor: %w", err)
	}

	// 4) Instantiate a binding to the ERC20 at rewardTokenAddr and read its decimals
	erc20, err := abigen.NewErc20(rewardTokenAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to ERC20: %w", err)
	}
	decimalsBig, err := erc20.Decimals(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals from ERC20: %w", err)
	}

	// Convert the decimals from uint8 to int64
	erc20Decimals := int64(decimalsBig)

	// 6) Assemble the result struct
	result := &syncedRewardData{
		Erc20Address:  rewardTokenAddr,
		Erc20Decimals: erc20Decimals,
	}

	return result, nil
}

// scaleUpUint256 turns a decimal into uint256, i.e. (11.22, 4) -> 112200
func scaleUpUint256(amount *types.Decimal, decimals uint16) (*types.Decimal, error) {
	unit, err := types.ParseDecimal("1" + strings.Repeat("0", int(decimals)))
	if err != nil {
		return nil, fmt.Errorf("create decimal unit failed: %w", err)
	}

	n, err := types.DecimalMul(amount, unit)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	err = n.SetPrecisionAndScale(uint256Precision, 0)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	return n, nil
}

// scaleDownUint256 turns an uint256 to a decimal, i.e. (112200, 4) -> 11.22
func scaleDownUint256(amount *types.Decimal, decimals uint16) (*types.Decimal, error) {
	unit, err := types.ParseDecimal("1" + strings.Repeat("0", int(decimals)))
	if err != nil {
		return nil, fmt.Errorf("create decimal unit failed: %w", err)
	}

	n, err := types.DecimalDiv(amount, unit)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	scale := n.Scale()
	if scale > decimals {
		scale = decimals
	}

	err = n.SetPrecisionAndScale(uint256Precision-decimals, scale)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	return n, nil
}

// parseAmountFromUser parses an amount from a user input string.
// It will scale the amount to the correct number of decimals.
func parseAmountFromUser(amount string, decimals uint16) (*types.Decimal, error) {
	dec, err := types.ParseDecimalExplicit(amount, 78, decimals)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	return scaleUpUint256(dec, decimals)
}
