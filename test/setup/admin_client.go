package setup

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/testcontainers/testcontainers-go"
)

var ErrTxNotConfirmed = errors.New("transaction not confirmed")

type AdminClient struct {
	container *testcontainers.DockerContainer
}
type cliResponse struct {
	Result any    `json:"result"`
	Error  string `json:"error"`
}

// exec runs a command in the admin container.
// It will JSON unmarshal the result into result.
func exec[T any](a *AdminClient, ctx context.Context, result T, args ...string) error {
	// request output in the json format
	args = append(args, "--output", "json")

	_, reader, err := a.container.Exec(ctx, append([]string{"/app/kwild"}, args...))
	if err != nil {
		return err
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	fmt.Println("Running Command ", `/app/kwild `+strings.Join(args, " "))
	fmt.Println("Exec output: ", string(output))

	// Find the first '{'
	startIdx := bytes.IndexByte(output, '{')
	if startIdx == -1 {
		return fmt.Errorf("no JSON object found in output")
	}

	// Trim everything before the first '{'
	trimmedOutput := output[startIdx:]
	resp := cliResponse{}
	err = json.Unmarshal(trimmedOutput, &resp)
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	bts, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}

	return json.Unmarshal(bts, result)
	// fmt.Println("Exec output: ", string(output))

	// d := display.MessageReader[T]{
	// 	Result: result,
	// }

	// err = json.NewDecoder(reader).Decode(&d)
	// if err != nil {
	// 	return err
	// }

	// if d.Error != "" {
	// 	return errors.New(d.Error)
	// }
}

func (a *AdminClient) ValidatorNodeJoin(ctx context.Context) (types.Hash, error) {
	var res display.TxHashResponse
	err := exec(a, ctx, &res, "validators", "join")
	if err != nil {
		return types.Hash{}, err
	}

	return res.TxHash, nil
}

func (a *AdminClient) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte, joinerKeyType crypto.KeyType) (types.Hash, error) {
	var res display.TxHashResponse
	pubKeyStr := config.EncodePubKeyAndType(joinerPubKey, joinerKeyType)
	err := exec(a, ctx, &res, "validators", "approve", pubKeyStr)
	if err != nil {
		return types.Hash{}, err
	}
	return res.TxHash, nil
}

func (a *AdminClient) ValidatorJoinStatus(ctx context.Context, joinerPubKey []byte, joinerKeyType crypto.KeyType) (*types.JoinRequest, error) {
	var res types.JoinRequest
	pubKeyStr := config.EncodePubKeyAndType(joinerPubKey, joinerKeyType)
	err := exec(a, ctx, &res, "validators", "join-status", pubKeyStr)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (a *AdminClient) ValidatorNodeLeave(ctx context.Context) (types.Hash, error) {
	var res display.TxHashResponse
	err := exec(a, ctx, &res, "validators", "leave")
	if err != nil {
		return types.Hash{}, err
	}

	return res.TxHash, nil
}

func (a *AdminClient) ValidatorNodeRemove(ctx context.Context, target []byte, pubKeyType crypto.KeyType) (types.Hash, error) {
	var res display.TxHashResponse
	keyStr := config.EncodePubKeyAndType(target, pubKeyType)
	err := exec(a, ctx, &res, "validators", "remove", keyStr)
	return res.TxHash, err
}

type valInfo struct {
	PubKey     string `json:"pubkey"`
	PubKeyType string `json:"pubkey_type"`
	Power      int64  `json:"power"`
}

func (a *AdminClient) ValidatorsList(ctx context.Context) ([]*types.Validator, error) {
	var res []*valInfo
	err := exec(a, ctx, &res, "validators", "list")
	if err != nil {
		return nil, err
	}

	validators := make([]*types.Validator, len(res))
	for i, v := range res {
		pubKey, err := hex.DecodeString(v.PubKey)
		if err != nil {
			return nil, err
		}

		pubKeyType, err := crypto.ParseKeyType(v.PubKeyType)
		if err != nil {
			return nil, err
		}

		validators[i] = &types.Validator{
			AccountID: types.AccountID{
				Identifier: pubKey,
				KeyType:    pubKeyType,
			},
			Power: v.Power,
		}
	}

	return validators, nil
}

func (a *AdminClient) TxSuccess(ctx context.Context, txHash types.Hash) error {
	var res *types.TxQueryResponse
	err := exec(a, ctx, &res, "utils", "query-tx", txHash.String())
	if err != nil {
		return err
	}

	// NOTE: this should not be considered a failure, should retry
	if res.Height < 0 {
		return ErrTxNotConfirmed
	}

	if res.Result != nil && res.Result.Code != 0 {
		return fmt.Errorf("tx failed: %v", res.Result)
	}

	return nil
}

func (a *AdminClient) SubmitMigrationProposal(ctx context.Context, activationHeight *big.Int, migrationDuration *big.Int) (types.Hash, error) {
	var res display.TxHashResponse
	err := exec(a, ctx, &res, "migrate", "propose", "-a", activationHeight.String(), "-d", migrationDuration.String())
	if err != nil {
		return types.Hash{}, err
	}
	return res.TxHash, nil
}

func (a *AdminClient) ApproveMigration(ctx context.Context, migrationResolutionID *types.UUID) (types.Hash, error) {
	var res display.TxHashResponse
	err := exec(a, ctx, &res, "migrate", "approve", migrationResolutionID.String())
	if err != nil {
		return types.Hash{}, err
	}
	return res.TxHash, nil
}

func (a *AdminClient) ListMigrations(ctx context.Context) ([]*types.Migration, error) {
	var res []*types.Migration
	err := exec(a, ctx, &res, "migrate", "proposals")
	if err != nil {
		return nil, err
	}
	return res, nil
}
