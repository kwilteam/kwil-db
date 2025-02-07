package erc20reward

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// this file implements a "named" erc20 extension, which is the extension that users interact with
func init() {
	err := precompiles.RegisterInitializer("erc20", func(ctx context.Context, service *common.Service,
		db sql.DB, alias string, metadata map[string]any) (precompiles.Precompile, error) {
		// the user can specify the following pieces of metadata:
		// - chain name (text) [required]
		// - escrow address (text) [required]
		// - distribution period (text) [optional]

		chainNameAny, ok := metadata["chain"]
		if !ok {
			return precompiles.Precompile{}, errors.New("missing required metadata field 'chain'")
		}
		chainName, ok := chainNameAny.(string)
		if !ok {
			return precompiles.Precompile{}, errors.New("metadata field 'chain' must be a string")
		}

		escrowAddressAny, ok := metadata["escrow"]
		if !ok {
			return precompiles.Precompile{}, errors.New("missing required metadata field 'escrow'")
		}

		escrowAddress, ok := escrowAddressAny.(string)
		if !ok {
			return precompiles.Precompile{}, errors.New("metadata field 'escrow' must be a string")
		}

		var distributionPeriod string
		distributionPeriodAny, ok := metadata["distribution_period"]
		if !ok {
			distributionPeriod = "1d"
		} else {
			distributionPeriod, ok = distributionPeriodAny.(string)
			if !ok {
				return precompiles.Precompile{}, errors.New("metadata field 'distribution_period' must be an int64")
			}
		}

		id := uuidForChainAndEscrow(chainName, escrowAddress)

		// makeMetaHandler makes a function that acts as a handler for calling methods on the meta extension.
		// It assumes the same function signature as the meta handler EXCEPT that the first argument is the id.
		makeMetaHandler := func(method string) precompiles.HandlerFunc {
			return func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
				_, err2 := app.Engine.Call(ctx, app.DB, RewardMetaExtensionName, method, append([]any{&id}, inputs...), func(r *common.Row) error {
					return resultFn(r.Values)
				})
				return err2
			}
		}

		return precompiles.Precompile{
			OnUse: func(ctx *common.EngineContext, app *common.App) error {
				id2, err := callPrepare(ctx, app, chainName, escrowAddress, distributionPeriod)
				if err != nil {
					return err
				}

				if *id2 != id {
					// indicates some basic error in the extension
					return errors.New("id mismatch")
				}

				return nil
			},
			OnUnuse: func(ctx *common.EngineContext, app *common.App) error {
				return callDisable(ctx, app, &id)
			},
			Methods: []precompiles.Method{
				{
					Name: "info",
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
					Handler:         makeMetaHandler("info"),
					AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
				},
				{
					Name: "id",
					Returns: &precompiles.MethodReturn{
						Fields: []precompiles.PrecompileValue{
							{Name: "id", Type: types.UUIDType},
						},
					},
					Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
						return resultFn([]any{id})
					},
					AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
				},
				{
					Name: "issue",
					Parameters: []precompiles.PrecompileValue{
						{Name: "user", Type: types.TextType},
						{Name: "amount", Type: types.TextType},
					},
					AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
					Handler:         makeMetaHandler("issue"),
				},
				{
					Name: "transfer",
					Parameters: []precompiles.PrecompileValue{
						{Name: "to", Type: types.TextType},
						{Name: "amount", Type: types.TextType},
					},
					// anybody can call this as long as they have the tokens.
					// There is no security risk if somebody calls this directly
					AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
					Handler:         makeMetaHandler("transfer"),
				},
				{
					Name: "lock",
					Parameters: []precompiles.PrecompileValue{
						{Name: "amount", Type: types.TextType},
					},
					AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
					Handler:         makeMetaHandler("lock"),
				},
				{
					Name: "lock_admin",
					Parameters: []precompiles.PrecompileValue{
						{Name: "user", Type: types.TextType},
						{Name: "amount", Type: types.TextType},
					},
					AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
					Handler:         makeMetaHandler("lock_admin"),
				},
				{
					Name: "unlock",
					Parameters: []precompiles.PrecompileValue{
						{Name: "user", Type: types.TextType},
						{Name: "amount", Type: types.TextType},
					},
					AccessModifiers: []precompiles.Modifier{precompiles.SYSTEM},
					Handler:         makeMetaHandler("unlock"),
				},
				{
					// balance returns the balance of a user.
					Name: "balance",
					Parameters: []precompiles.PrecompileValue{
						{Name: "user", Type: types.TextType},
					},
					Returns: &precompiles.MethodReturn{
						Fields: []precompiles.PrecompileValue{
							{Name: "balance", Type: types.TextType},
						},
					},
					AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
					Handler:         makeMetaHandler("balance"),
				},
			},
		}, nil
	})
	if err != nil {
		panic(err)
	}
}
