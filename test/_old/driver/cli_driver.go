package driver

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ec "github.com/ethereum/go-ethereum/crypto"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/database"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"

	ethdeployer "github.com/kwilteam/kwil-db/test/integration/eth-deployer"
)

// KwilCliDriver is a driver for tests using `cmd/kwil-cli`
type KwilCliDriver struct {
	cliBin          string // kwil-cli binary path
	rpcURL          string
	privKey         string
	identity        []byte
	gatewayProvider bool
	chainID         string
	deployer        *ethdeployer.Deployer
	logger          log.Logger
}

func NewKwilCliDriver(cliBin, rpcURL, privKey, chainID string, identity []byte, gatewayProvider bool, deployer *ethdeployer.Deployer, logger log.Logger) *KwilCliDriver {
	return &KwilCliDriver{
		cliBin:          cliBin,
		rpcURL:          rpcURL,
		privKey:         privKey,
		identity:        identity,
		gatewayProvider: gatewayProvider,
		logger:          logger,
		chainID:         chainID,
		deployer:        deployer,
	}
}

// newKwilCliCmd returns a new exec.Cmd for kwil-cli
func (d *KwilCliDriver) newKwilCliCmd(args ...string) *exec.Cmd {
	args = append(args, "--provider", d.rpcURL)
	args = append(args, "--private-key", d.privKey)
	args = append(args, "--chain-id", d.chainID)
	args = append(args, "--output", "json")

	d.logger.Info("cli Cmd", "args",
		strings.Join(append([]string{d.cliBin}, args...), " "))

	cmd := exec.Command(d.cliBin, args...)
	return cmd
}

// newKwilCliCmdWithYes this is a helper function to automatically answer yes to
// all prompts. This is useful for testing.
// The cmd will be executed as `yes | kwil-cli <args>`
func (d *KwilCliDriver) newKwilCliCmdWithYes(args ...string) *exec.Cmd {
	args = append([]string{"yes |", d.cliBin}, args...)

	args = append(args, "--provider", d.rpcURL)
	args = append(args, "--private-key", d.privKey)
	args = append(args, "--chain-id", d.chainID)
	args = append(args, "--output", "json")

	s := strings.Join(args, " ")

	d.logger.Info("cli Cmd(with yes)", "args",
		strings.Join(append([]string{"bash", "-c"}, s), " "))

	cmd := exec.Command("bash", "-c", s)
	return cmd
}

// SupportBatch
func (d *KwilCliDriver) SupportBatch() bool {
	return false
}

func (d *KwilCliDriver) account(_ context.Context, acctID []byte) (*types.Account, error) {
	cmd := d.newKwilCliCmd("account", "balance", hex.EncodeToString(acctID))
	out, err := mustRun[respAccount](cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", err)
	}

	d.logger.Debug("account balance result", "result", out)
	acct, err := parseRespAccount(out)
	if err != nil {
		return nil, fmt.Errorf("failed to parse account balance response: %w", err)
	}

	return acct, nil
}

func (d *KwilCliDriver) AccountBalance(ctx context.Context, acctID []byte) (*big.Int, error) {
	acct, err := d.account(ctx, acctID)
	if err != nil {
		return nil, err
	}

	return acct.Balance, nil
}

func (d *KwilCliDriver) TransferAmt(ctx context.Context, to []byte, amt *big.Int) (types.Hash, error) {
	cmd := d.newKwilCliCmd("account", "transfer", hex.EncodeToString(to), amt.String())
	out, err := mustRun[respTxHash](cmd, d.logger)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to do acct transfer: %w", err)
	}

	return out.TxHash, nil
}

func (d *KwilCliDriver) TxSuccess(_ context.Context, txHash types.Hash) error {
	cmd := d.newKwilCliCmd("utils", "query-tx", txHash.String())
	out, err := mustRun[respTxQuery](cmd, d.logger)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrTxNotConfirmed // not quite, but for this driver it's a retry condition
		}
		return fmt.Errorf("failed to query tx: %w", err)
	}

	d.logger.Debug("tx info", "height", out.Height,
		"txHash", txHash,
		"result", out.TxResult)

	// NOTE: this should not be considered a failure, should retry
	if out.Height < 0 {
		return ErrTxNotConfirmed
	}

	if out.TxResult.Code != 0 {
		return fmt.Errorf("tx failed: %s", out.TxResult.Log)
	}
	return nil
}

func (d *KwilCliDriver) DropDatabase(_ context.Context, dbName string) (types.Hash, error) {
	cmd := d.newKwilCliCmd("database", "drop", dbName)
	out, err := mustRun[respTxHash](cmd, d.logger)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to drop database: %w", err)
	}

	d.logger.Debug("drop database tx", "result", out.TxHash)

	return out.TxHash, nil
}

// prepareCliActionParams returns the named action args for the given action name, in
// the format of `name:value`
func (d *KwilCliDriver) prepareCliActionParams(ctx context.Context, namespace string, actionName string, actionInputs []any) ([]string, error) {
	params, err := database.GetParamList(ctx, func(ctx context.Context, query string, args map[string]any) (*types.QueryResult, error) {
		return d.QueryDatabase(ctx, query)
	}, namespace, actionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get action params: %w", err)
	}

	if len(params) != len(actionInputs) {
		return nil, fmt.Errorf("invalid number of inputs, expected %d, got %d", len(params), len(actionInputs))
	}

	stringify := func(v any) string {
		switch v := v.(type) {
		case []byte:
			return base64.StdEncoding.EncodeToString(v) + "#b64"
		case fmt.Stringer:
			return v.String()
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	args := []string{}
	for i, input := range params {
		pName := input.Name[1:] // remove the leading $

		// if the input is a slice, we need to delimit it with commas
		typeOf := reflect.TypeOf(actionInputs[i])
		if typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() != reflect.Uint8 {
			var sliceArgs []string
			for _, v := range actionInputs[i].([]any) {
				sliceArgs = append(sliceArgs, stringify(v))
			}
			args = append(args, fmt.Sprintf("%s:%s", pName, strings.Join(sliceArgs, ",")))
			continue
		}

		args = append(args, fmt.Sprintf("%s:%s", pName, stringify(actionInputs[i])))
	}
	return args, nil
}

func (d *KwilCliDriver) Execute(ctx context.Context, namespace string, action string, inputs ...[]any) (types.Hash, error) {
	if len(inputs) > 1 {
		return types.Hash{}, fmt.Errorf("kwil-cli does not support batched inputs")
	}

	// NOTE: kwil-cli does not support batched inputs
	actionInputs, err := d.prepareCliActionParams(ctx, namespace, action, inputs[0])
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to get action params: %w", err)
	}

	args := []string{"database", "execute", action, "--namespace", namespace}
	args = append(args, actionInputs...)

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun[respTxHash](cmd, d.logger)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to execute action: %w", err)
	}
	return out.TxHash, nil
}

func (d *KwilCliDriver) ExecuteSQL(ctx context.Context, sql string, params map[string]any) (types.Hash, error) {
	// actionInputs, err := d.prepareCliActionParams(ctx, dbid, action, inputs[0])
	// if err != nil {
	// 	return types.Hash{}, fmt.Errorf("failed to get action params: %w", err)
	// }

	args := []string{"database", "execute", "--sql", sql}

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun[respTxHash](cmd, d.logger)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to execute action: %w", err)
	}
	return out.TxHash, nil
}

func (d *KwilCliDriver) QueryDatabase(_ context.Context, query string) (*types.QueryResult, error) {
	args := []string{"database", "query", query}

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun[*types.QueryResult](cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	d.logger.Debug("query result", "result", out)
	return out, nil
}

func (d *KwilCliDriver) Call(ctx context.Context, dbid, action string, inputs []any) (*types.CallResult, error) {
	// NOTE: kwil-cli does not support batched inputs
	actionInputs, err := d.prepareCliActionParams(ctx, dbid, action, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare action params: %w", err)
	}

	args := []string{"database", "call", action, "--dbid", dbid, "--logs"}
	args = append(args, actionInputs...)

	if d.gatewayProvider {
		args = append(args, "--authenticate")
	}

	cmd := d.newKwilCliCmdWithYes(args...)

	out, err := mustRunCallIgnorePrompt[*types.CallResult](cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to call action: %w", err)
	}
	d.logger.Debug("call result", "result", out)

	return out, nil
}

func (d *KwilCliDriver) ChainInfo(_ context.Context) (*types.ChainInfo, error) {
	cmd := d.newKwilCliCmd("utils", "chain-info")
	out, err := mustRun[*types.ChainInfo](cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain info: %w", err)
	}

	d.logger.Debug("chain info", "Resp", out)

	return out, nil
}

///////// helper functions

type genericResponse struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// mustRun runs the give command, and parse stdout
func mustRun[T any](cmd *exec.Cmd, logger log.Logger) (T, error) {
	cmd.Stderr = os.Stderr
	var t T
	// here we capture the stdout
	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		return t, err
	}

	output := out.Bytes()
	logger.Debug("cmd output", "output", string(output))

	var jsonResult genericResponse
	err = json.Unmarshal(output, &jsonResult)
	if err != nil {
		logger.Error("bad cmd output", "error", err, "output", string(output), "stderr", stdErr.String())
		return t, err
	}

	if jsonResult.Error != "" {
		return t, errors.New(jsonResult.Error)
	}

	err = json.Unmarshal(jsonResult.Result, &t)
	if err != nil {
		logger.Error("bad cmd output result field", err, "result", string(jsonResult.Result), "stderr", stdErr.String())
		return t, err
	}

	return t, nil
}

// mustRunCallIgnorePrompt runs the given `kwil-cli database call` command, and
// throw away the prompt output. This is necessary for authn call, because
// kwil-cli will prompt for confirmation.
func mustRunCallIgnorePrompt[T any](cmd *exec.Cmd, logger log.Logger) (T, error) {
	cmd.Stderr = os.Stderr
	var t T
	// here we capture the stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return t, err
	}

	output := out.Bytes()
	// logger.Debug("cmd output", "output", string(output))

	// This is a bit hacky, throw away the first part prompt output, if any
	prompted := "Do you want to sign this message?"
	delimiter := "{\n"
	if strings.Contains(string(output), prompted) {
		logger.Debug("throw away prompt output")
		output = []byte(delimiter + strings.SplitN(string(output), delimiter, 2)[1])
	}

	var jsonResult genericResponse
	err = json.Unmarshal(output, &jsonResult)
	if err != nil {
		logger.Error("bad cmd output", "error", err, "output", string(output))
		return t, err
	}

	if jsonResult.Error != "" {
		return t, errors.New(jsonResult.Error)
	}

	err = json.Unmarshal(jsonResult.Result, &t)
	if err != nil {
		logger.Error("bad cmd output result field", "error", err, "result", string(jsonResult.Result))
		return t, err
	}

	return t, nil
}

// NOTE: trivial to implement. Another way is to import resp* structure
// from cmd/kwil-cli/cmds,
type respTxHash struct {
	TxHash types.Hash `json:"tx_hash"`
}

// respTxQuery represents the tx query response(json) from the cli response
type respTxQuery struct {
	Height   int64 `json:"height"`
	TxResult struct {
		Code uint32 `json:"code"`
		Log  string `json:"log"`
	} `json:"tx_result"`
}

type respAccount struct {
	Identifier string `json:"identifier"`
	Balance    string `json:"balance"`
	Nonce      int64  `json:"nonce"`
}

func parseRespAccount(data respAccount) (*types.Account, error) {
	acctID, err := hex.DecodeString(data.Identifier)
	if err != nil {
		return nil, fmt.Errorf("invalid identifier hex string: %w", err)
	}

	bal, ok := big.NewInt(0).SetString(data.Balance, 10)
	if !ok {
		return nil, errors.New("invalid decimal string balance")
	}

	acct := &types.Account{
		Identifier: acctID,
		Balance:    bal,
		Nonce:      data.Nonce,
	}
	return acct, nil
}

func (d *KwilCliDriver) Approve(ctx context.Context, sender *ecdsa.PrivateKey, amount *big.Int) error {
	if d.deployer == nil {
		return fmt.Errorf("deployer is nil")
	}

	return d.deployer.Approve(ctx, sender, amount)
}

func (d *KwilCliDriver) Deposit(ctx context.Context, sender *ecdsa.PrivateKey, amount *big.Int) error {
	if d.deployer == nil {
		return fmt.Errorf("deployer is nil")
	}

	return d.deployer.Deposit(ctx, sender, amount)
}

func (d *KwilCliDriver) EscrowBalance(ctx context.Context, senderPrivKeyECDSA *ecdsa.PrivateKey) (*big.Int, error) {
	if d.deployer == nil {
		return nil, fmt.Errorf("deployer is nil")
	}

	priv, err := crypto.UnmarshalSecp256k1PrivateKey(senderPrivKeyECDSA.D.Bytes())
	if err != nil {
		return nil, err
	}

	senderAddr := crypto.EthereumAddressFromPubKey(priv.Public().(*crypto.Secp256k1PublicKey))
	var ethAddr common.Address
	copy(ethAddr[:], senderAddr)

	return d.deployer.EscrowBalance(ctx, ethAddr)
}

func (d *KwilCliDriver) UserBalance(ctx context.Context, senderAddress *ecdsa.PrivateKey) (*big.Int, error) {
	if d.deployer == nil {
		return nil, fmt.Errorf("deployer is nil")
	}

	senderAddr := ec.PubkeyToAddress(senderAddress.PublicKey)
	return d.deployer.UserBalance(ctx, senderAddr)
}

func (d *KwilCliDriver) Allowance(ctx context.Context, sender *ecdsa.PrivateKey) (*big.Int, error) {
	if d.deployer == nil {
		return nil, fmt.Errorf("deployer is nil")
	}

	senderAddr := ec.PubkeyToAddress(sender.PublicKey)
	return d.deployer.Allowance(ctx, senderAddr)
}

func (d *KwilCliDriver) Signer() []byte {
	return d.identity
}

func (d *KwilCliDriver) Identifier() (string, error) {
	return auth.EthSecp256k1Authenticator{}.Identifier(d.Signer())
}

func (d *KwilCliDriver) TxInfo(ctx context.Context, hash types.Hash) (*types.TxQueryResponse, error) {

	args := []string{"utils", "query-tx", hash.String(), "--full"}

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun[*types.TxQueryResponse](cmd, d.logger)
	if err != nil {
		if strings.Contains(err.Error(), "transaction not found") {
			// try again, hacking around comet's mempool inconsistency
			time.Sleep(500 * time.Millisecond)
			res2, err := mustRun[*types.TxQueryResponse](cmd, d.logger)
			if err != nil {
				return nil, err
			}
			return res2, nil
		}
		return nil, err
	}

	return out, nil
}
