package driver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"

	"go.uber.org/zap"
)

// KwilCliDriver is a driver for tests using `cmd/kwil-cli`
type KwilCliDriver struct {
	cliBin          string // kwil-cli binary path
	rpcUrl          string
	privKey         string
	identity        []byte
	gatewayProvider bool
	chainID         string
	logger          log.Logger
}

func NewKwilCliDriver(cliBin, rpcUrl, privKey, chainID string, identity []byte, gatewayProvider bool, logger log.Logger) *KwilCliDriver {
	return &KwilCliDriver{
		cliBin:          cliBin,
		rpcUrl:          rpcUrl,
		privKey:         privKey,
		identity:        identity,
		gatewayProvider: gatewayProvider,
		logger:          logger,
		chainID:         chainID,
	}
}

// newKwilCliCmd returns a new exec.Cmd for kwil-cli
func (d *KwilCliDriver) newKwilCliCmd(args ...string) *exec.Cmd {
	args = append(args, "--kwil-provider", d.rpcUrl)
	args = append(args, "--private-key", d.privKey)
	args = append(args, "--chain-id", d.chainID)
	args = append(args, "--output", "json")

	d.logger.Info("cli Cmd", zap.String("args",
		strings.Join(append([]string{d.cliBin}, args...), " ")))

	cmd := exec.Command(d.cliBin, args...)
	return cmd
}

// newKwilCliCmdWithYes this is a helper function to automatically answer yes to
// all prompts. This is useful for testing.
// The cmd will be executed as `yes | kwil-cli <args>`
func (d *KwilCliDriver) newKwilCliCmdWithYes(args ...string) *exec.Cmd {
	args = append([]string{"yes |", d.cliBin}, args...)

	args = append(args, "--kwil-provider", d.rpcUrl)
	args = append(args, "--private-key", d.privKey)
	args = append(args, "--chain-id", d.chainID)
	args = append(args, "--output", "json")

	s := strings.Join(args, " ")

	d.logger.Info("cli Cmd(with yes)", zap.String("args",
		strings.Join(append([]string{"bash", "-c"}, s), " ")))

	cmd := exec.Command("bash", "-c", s)
	return cmd
}

// SupportBatch
func (d *KwilCliDriver) SupportBatch() bool {
	return false
}

func (d *KwilCliDriver) account(ctx context.Context, acctID []byte) (*types.Account, error) {
	cmd := d.newKwilCliCmd("account", "balance", hex.EncodeToString(acctID))
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	d.logger.Debug("account balance result", zap.Any("result", out.Result))
	acct, err := parseRespAccount(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list databases response: %w", err)
	}

	return acct, nil
}

func (d *KwilCliDriver) AccountBalance(ctx context.Context, acctID []byte) (*big.Int, error) {
	acct, err := d.account(ctx, acctID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list databases response: %w", err)
	}

	return acct.Balance, nil
}

func (d *KwilCliDriver) TransferAmt(ctx context.Context, to []byte, amt *big.Int) (txHash []byte, err error) {
	cmd := d.newKwilCliCmd("account", "transfer", hex.EncodeToString(to), amt.String())
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	txHash, err = parseRespTxHash(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) DBID(name string) string {
	return utils.GenerateDBID(name, d.identity)
}

func (d *KwilCliDriver) listDatabase() ([]*types.DatasetIdentifier, error) {
	cmd := d.newKwilCliCmd("database", "list", "--owner", hex.EncodeToString(d.identity))
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	d.logger.Debug("list database result", zap.Any("result", out.Result))
	dbs, err := parseRespListDatabases(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse list databases response: %w", err)
	}

	return dbs, nil
}

func (d *KwilCliDriver) DatabaseExists(_ context.Context, dbid string) error {
	// check GetSchema
	_, err := d.getSchema(dbid)
	if err != nil {
		return err
	}

	// check ListDatabases
	dbs, err := d.listDatabase()
	if err != nil {
		return err
	}

	found := false
	for _, db := range dbs {
		if db.DBID == dbid {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("ListDatabase: database not found: %s", dbid)
	}

	return nil
}

func (d *KwilCliDriver) DeployDatabase(_ context.Context, db *transactions.Schema) (txHash []byte, err error) {
	schemaFile := path.Join(os.TempDir(), fmt.Sprintf("schema-%s.json", time.Now().Format("20060102150405")))

	dbByte, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal database: %w", err)
	}

	err = os.WriteFile(schemaFile, dbByte, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write database schema: %w", err)
	}

	cmd := d.newKwilCliCmd("database", "deploy", "-p", schemaFile, "-t", "json")
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy databse: %w", err)
	}

	txHash, err = parseRespTxHash(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) TxSuccess(_ context.Context, txHash []byte) error {
	cmd := d.newKwilCliCmd("utils", "query-tx", hex.EncodeToString(txHash))
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrTxNotConfirmed // not quite, but for this driver it's a retry condition
		}
		return fmt.Errorf("failed to query tx: %w", err)
	}

	resp, err := parseRespTxQuery(out.Result)
	if err != nil {
		d.logger.Debug("tx query failed", zap.String("error", err.Error()))
		return fmt.Errorf("query failed: %w", err)
	}

	d.logger.Debug("tx info", zap.Int64("height", resp.Height),
		zap.String("txHash", hex.EncodeToString(txHash)),
		zap.Any("result", resp.TxResult))

	// NOTE: this should not be considered a failure, should retry
	if resp.Height < 0 {
		return ErrTxNotConfirmed
	}

	if resp.TxResult.Code != 0 {
		return fmt.Errorf("tx failed: %s", resp.TxResult.Log)
	}
	return nil
}

func (d *KwilCliDriver) DropDatabase(_ context.Context, dbName string) (txHash []byte, err error) {
	cmd := d.newKwilCliCmd("database", "drop", dbName)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to drop database: %w", err)
	}

	d.logger.Debug("drop database tx", zap.Any("result", out.Result))
	txHash, err = parseRespTxHash(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) getSchema(dbid string) (*transactions.Schema, error) {
	cmd := d.newKwilCliCmd("database", "read-schema", "--dbid", dbid)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to getSchema: %w", err)
	}

	schema, err := parseRespGetSchema(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse getSchema response: %w", err)
	}
	return schema, nil
}

// prepareCliActionParams returns the named action args for the given action name, in
// the format of `name:value`
func (d *KwilCliDriver) prepareCliActionParams(dbid string, actionName string, actionInputs []any) ([]string, error) {
	schema, err := d.getSchema(dbid)
	if err != nil {
		return nil, err
	}

	var action *transactions.Action
	for _, a := range schema.Actions {
		if a.Name == actionName {
			action = a
			break
		}
	}

	if len(action.Inputs) != len(actionInputs) {
		return nil, fmt.Errorf("invalid number of inputs, expected %d, got %d", len(action.Inputs), len(actionInputs))
	}

	args := []string{}
	for i, input := range action.Inputs {
		input = input[1:] // remove the leading $
		args = append(args, fmt.Sprintf("%s:%v", input, actionInputs[i]))
	}
	return args, nil
}

func (d *KwilCliDriver) ExecuteAction(_ context.Context, dbid string, action string, inputs ...[]any) ([]byte, error) {
	// NOTE: kwil-cli does not support batched inputs
	actionInputs, err := d.prepareCliActionParams(dbid, action, inputs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get action params: %w", err)
	}

	args := []string{"database", "execute", "--dbid", dbid, "--action", action}
	args = append(args, actionInputs...)

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to execute action: %w", err)
	}

	txHash, err := parseRespTxHash(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) QueryDatabase(_ context.Context, dbid, query string) (*client.Records, error) {
	cmd := d.newKwilCliCmd("database", "query", "--dbid", dbid, query)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	d.logger.Debug("query result", zap.Any("result", out.Result))
	records, err := parseRespQueryDb(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}
	return records, nil
}

func (d *KwilCliDriver) Call(_ context.Context, dbid, action string, inputs []any) (*client.Records, error) {
	// NOTE: kwil-cli does not support batched inputs
	actionInputs, err := d.prepareCliActionParams(dbid, action, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare action params: %w", err)
	}

	args := []string{"database", "call", "--dbid", dbid, "--action", action}
	args = append(args, actionInputs...)

	if d.gatewayProvider {
		args = append(args, "--authenticate")
	}

	cmd := d.newKwilCliCmdWithYes(args...)
	out, err := mustRunCallIgnorePrompt(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to call action: %w", err)
	}
	d.logger.Debug("call result", zap.Any("result", out.Result))

	return parseRespQueryDb(out.Result)
}

func (d *KwilCliDriver) ChainInfo(_ context.Context) (*types.ChainInfo, error) {
	cmd := d.newKwilCliCmd("utils", "chain-info")
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain info: %w", err)
	}

	d.logger.Debug("chain info", zap.Any("Resp", out.Result))
	var chainInfo types.ChainInfo

	bts, err := json.Marshal(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chain info: %w", err)
	}

	err = json.Unmarshal(bts, &chainInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain info: %w", err)
	}

	return &chainInfo, nil
}

///////// helper functions

// mustRun runs the give command, and parse stdout
func mustRun(cmd *exec.Cmd, logger log.Logger) (*cliResponse, error) {
	cmd.Stderr = os.Stderr
	//// here we capture the stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	output := out.Bytes()
	// logger.Debug("cmd output", zap.String("output", string(output)))

	var result *cliResponse
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

// mustRunCallIgnorePrompt runs the given `kwil-cli database call` command, and
// throw away the prompt output. This is necessary for authn call, because
// kwil-cli will prompt for confirmation.
func mustRunCallIgnorePrompt(cmd *exec.Cmd, logger log.Logger) (*cliResponse, error) {
	cmd.Stderr = os.Stderr
	//// here we capture the stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	output := out.Bytes()
	// logger.Debug("cmd output", zap.String("output", string(output)))

	// This is a bit hacky, throw away the first part prompt output, if any
	prompted := "Do you want to sign this message?"
	delimiter := "{\n"
	if strings.Contains(string(output), prompted) {
		logger.Debug("throw away prompt output")
		output = []byte(delimiter + strings.SplitN(string(output), delimiter, 2)[1])
	}

	var result *cliResponse
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

type cliResponse struct {
	Result any    `json:"result"` // json.RawMessage
	Error  string `json:"error"`
}

// Types below (resp*) are kind of duplicated with `cmd/kwil-cli`,  and i probably
// should expose those types from `cmd/kwil-cli` and use the
// `encoding.TextUnmarshaler` interface. thus enables unit testing
// Why i didn't do that is because:
// - for this driver, we only need to parse few types in `cmd/kwil-cli`
//
// If we are going to mock test for kwil-cli, we should do that.

// NOTE: trivial to implement. Another way is to import resp* structure
// from cmd/kwil-cli/cmds,
type respTxHash struct {
	TxHash string `json:"tx_hash"`
}

// parseRespTxHash parses the tx hash response(json) from the cli response
func parseRespTxHash(data any) ([]byte, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tx resp: %w", err)
	}

	var resp respTxHash
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tx hash: %w", err)
	}

	txHash, err := hex.DecodeString(resp.TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tx hash: %w", err)
	}

	return txHash, nil
}

// respTxQuery represents the tx query response(json) from the cli response
type respTxQuery struct {
	Height   int64 `json:"height"`
	TxResult struct {
		Code uint32 `json:"code"`
		Log  string `json:"log"`
	} `json:"tx_result"`
}

// parserRespTxQuery parses the tx query response(json) from the cli response
func parseRespTxQuery(data any) (*respTxQuery, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tx query resp: %w", err)
	}

	var resp respTxQuery
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tx query: %w", err)
	}

	return &resp, nil
}

// parseRespGetSchema parses the get schema response(json) from the cli response
func parseRespGetSchema(data any) (*transactions.Schema, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get schema resp: %w", err)
	}

	var resp transactions.Schema
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal get sceham: %w", err)
	}

	return &resp, nil
}

// respQueryDb represents the query db response(json) from the cli response
type respQueryDb []map[string]any

func parseRespQueryDb(data any) (*client.Records, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query db resp: %w", err)
	}

	var resp respQueryDb
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal query db: %w", err)
	}

	return client.NewRecordsFromMaps(resp), nil
}

func parseRespListDatabases(data any) ([]*types.DatasetIdentifier, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal list databases resp: %w", err)
	}

	var resp []*types.DatasetIdentifier
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal list databases: %w", err)
	}

	return resp, nil
}

type respAccount struct {
	Identifier string `json:"identifier"`
	Balance    string `json:"balance"`
	Nonce      int64  `json:"nonce"`
}

func parseRespAccount(data any) (*types.Account, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal list databases resp: %w", err)
	}

	var resp respAccount
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal list databases: %w", err)
	}

	acctID, err := hex.DecodeString(resp.Identifier)
	if err != nil {
		return nil, fmt.Errorf("invalid identifier hex string: %w", err)
	}

	bal, ok := big.NewInt(0).SetString(resp.Balance, 10)
	if !ok {
		return nil, errors.New("invalid decimal string balance")
	}

	acct := &types.Account{
		Identifier: acctID,
		Balance:    bal,
		Nonce:      resp.Nonce,
	}
	return acct, nil
}
