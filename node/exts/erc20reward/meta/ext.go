// Package meta is the reward_meta extension. All tables reside in meta_extension namespace.
// So in future if we can support other smart contract platforms.
package meta

import (
	"context"
	"fmt"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"

	kcommon "github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	pc "github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

func init() {
	// erc20rw_rewards_meta will be available to be used in the Kuneiform schema
	// It's supposed to be used by other reward extensions.
	err := pc.RegisterInitializer("erc20_rewards_meta",
		func(ctx context.Context, service *kcommon.Service, db sql.DB, alias string, metadata map[string]any) (pc.Precompile, error) {
			if alias != ExtAlias { // ensure only ExtAlias can be used
				return pc.Precompile{}, fmt.Errorf("invalid alias")
			}

			ext := &Erc20RewardMeta{alias: alias}

			methods := []pc.Method{
				{
					Name:            "register",
					AccessModifiers: []pc.Modifier{pc.SYSTEM},
					Handler: func(ctx *kcommon.EngineContext, app *kcommon.App, inputs []any, resultFn func([]any) error) error {
						return ext.register(ctx, app, inputs, resultFn)
					},
					Parameters: []pc.PrecompileValue{
						{Type: types.IntType, Nullable: false},
						{Type: types.TextType, Nullable: false},
						{Type: types.IntType, Nullable: false},
						{Type: types.TextType, Nullable: false},
						//{Type: types.TextArrayType, Nullable: false},
						{Type: types.IntType, Nullable: false},
						{Type: types.TextType, Nullable: false},
						{Type: types.IntType, Nullable: false},
					},
				},
				{
					Name:            "unregister",
					AccessModifiers: []pc.Modifier{pc.SYSTEM},
					Handler: func(ctx *kcommon.EngineContext, app *kcommon.App, inputs []any, resultFn func([]any) error) error {
						return ext.unregister(ctx, app, inputs, resultFn)
					},
				},
			}

			return pc.Precompile{
				OnStart: nil,
				Methods: methods,
				OnUse: func(ctx *kcommon.EngineContext, app *kcommon.App) error {
					ctx.OverrideAuthz = true
					defer func() { ctx.OverrideAuthz = false }()

					err := app.Engine.Execute(ctx, app.DB, sqlInitTableErc20rwContracts, nil, nil)
					if err != nil {
						return err
					}

					err = app.Engine.Execute(ctx, app.DB, sqlInitTableErc20rwSigners, nil, nil)
					return err
				},
				OnUnuse: func(ctx *kcommon.EngineContext, app *kcommon.App) error {
					return nil
				},
			}, nil
		})

	if err != nil {
		panic(fmt.Errorf("failed to register erc20rw_rewards_meta extension: %w", err))
	}
}

type Erc20RewardMeta struct {
	alias string
}

func (h *Erc20RewardMeta) register(ctx *kcommon.EngineContext, app *kcommon.App, inputs []any, resultFn func([]any) error) error {
	if len(inputs) != 7 {
		return fmt.Errorf("internal bud: expect 7 params, got %d", len(inputs))
	}

	chainID, ok := inputs[0].(int64)
	if !ok {
		return fmt.Errorf("invalid chain_id")
	}

	contractAddress, ok := inputs[1].(string)
	if !ok {
		return fmt.Errorf("invalid contract_address")
	}

	contractNonce, ok := inputs[2].(int64)
	if !ok {
		return fmt.Errorf("invalid contract_nonce")
	}

	if !ethCommon.IsHexAddress(contractAddress) {
		return fmt.Errorf("invalid contract_address")
	}

	signersStr, ok := inputs[3].(string)
	if !ok {
		return fmt.Errorf("invalid signers")
	}

	if len(signersStr) == 0 {
		return fmt.Errorf("signers is empty")
	}
	signers := strings.Split(signersStr, ",")
	for _, signer := range signers {
		if !ethCommon.IsHexAddress(signer) {
			return fmt.Errorf("invalid signer")
		}
	}

	threshold, ok := inputs[4].(int64)
	if !ok {
		return fmt.Errorf("invalid threshold")
	}

	if threshold == 0 {
		return fmt.Errorf("threshold is 0")
	}

	if threshold > int64(len(signers)) {
		return fmt.Errorf("threshold is larger than the number of signers")
	}

	safeAddress, ok := inputs[5].(string)
	if !ok {
		return fmt.Errorf("invalid safe_address")
	}

	if !ethCommon.IsHexAddress(safeAddress) {
		return fmt.Errorf("invalid safe_address")
	}

	safeNonce, ok := inputs[6].(int64)
	if !ok {
		return fmt.Errorf("invalid safe_nonce")
	}

	contractID := GenRewardContractID(chainID, contractAddress)

	// initialy, this is the params for  sqlCreateRewardContract
	combinedParams := map[string]any{
		//"$contract_id":  contractID[:], seems QueryPlanner enforces the type check on uuid, so we cannot use [:]
		"$contract_id":  contractID,
		"$chain_id":     chainID,
		"$address":      contractAddress,
		"$nonce":        contractNonce,
		"$threshold":    threshold,
		"$safe_address": safeAddress,
		"$safe_nonce":   safeNonce,
	}

	createSignerSql := fmt.Sprintf(`{%s}INSERT INTO erc20rw_meta_signers (id, address, contract_id) VALUES `, ExtAlias)
	for i, signer := range signers {
		if i > 0 {
			createSignerSql += ","
		}
		createSignerSql += fmt.Sprintf("($signer_id%d, $signer%d, $contract_id)", i, i)
		combinedParams[fmt.Sprintf("$signer_id%d", i)] = GenSignerID(contractID, signer)
		combinedParams[fmt.Sprintf("$signer%d", i)] = signer
	}
	createSignerSql += ";"

	ctx.OverrideAuthz = true
	defer func() { ctx.OverrideAuthz = false }()
	err := app.Engine.Execute(ctx, app.DB, sqlCreateRewardContract+createSignerSql, combinedParams, nil)
	if err != nil {
		return err
	}

	return nil
}

func (h *Erc20RewardMeta) unregister(ctx *kcommon.EngineContext, app *kcommon.App, inputs []any, resultFn func([]any) error) error {
	return nil
}

// Map turns a []T1 to a []T2 using a mapping function.
func Map[T1, T2 any](s []T1, f func(T1) T2) []T2 {
	r := make([]T2, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
