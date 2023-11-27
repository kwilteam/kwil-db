package driver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	cliBin     string // kwil-cli binary path
	adminBin   string // kwil-admin binary path
	rpcUrl     string
	privKey    string
	identifier []byte
	logger     log.Logger
}

func NewKwilCliDriver(cliBin, adminBin, rpcUrl, privKey string, pubKey []byte, logger log.Logger) *KwilCliDriver {
	return &KwilCliDriver{
		cliBin:     cliBin,
		adminBin:   adminBin,
		rpcUrl:     rpcUrl,
		privKey:    privKey,
		identifier: pubKey,
		logger:     logger,
	}
}

func (d *KwilCliDriver) newKwilCliCmd(args ...string) *exec.Cmd {
	args = append(args, "--kwil-provider", d.rpcUrl)
	args = append(args, "--private-key", d.privKey)
	args = append(args, "--output", "json")

	d.logger.Info("cli Cmd", zap.String("args",
		strings.Join(append([]string{d.cliBin}, args...), " ")))

	cmd := exec.Command(d.cliBin, args...)
	return cmd
}

func (d *KwilCliDriver) newKwilAdminCmd(args ...string) *exec.Cmd {
	args = append(args, "--rpcserver", d.rpcUrl)
	args = append(args, "--output", "json")

	d.logger.Info("admin cmd", zap.String("args",
		strings.Join(append([]string{d.adminBin}, args...), " ")))

	cmd := exec.Command(d.adminBin, args...)
	return cmd
}

// SupportBatch
// kwil-cli does not support batched inputs.
func (d *KwilCliDriver) SupportBatch() bool {
	return false
}

func (d *KwilCliDriver) DBID(name string) string {
	return utils.GenerateDBID(name, d.identifier)
}

func (d *KwilCliDriver) DatabaseExists(_ context.Context, dbid string) error {
	cmd := d.newKwilCliCmd("database", "read-schema", "--dbid", dbid)
	_, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to read database schema", zap.Error(err))
		return fmt.Errorf("failed to read database schema: %w", err)
	}

	return nil
}

func (d *KwilCliDriver) DeployDatabase(_ context.Context, db *transactions.Schema) (txHash []byte, err error) {
	schemaFile := path.Join(os.TempDir(), fmt.Sprintf("schema-%s.json", time.Now().Format("20060102150405")))

	dbByte, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		d.logger.Error("failed to marshal database", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal database: %w", err)
	}

	err = os.WriteFile(schemaFile, dbByte, 0644)
	if err != nil {
		d.logger.Error("failed to write database schema", zap.Error(err))
		return nil, fmt.Errorf("failed to write database schema: %w", err)
	}

	cmd := d.newKwilCliCmd("database", "deploy", "-p", schemaFile, "-t", "json")
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Warn("db schema", zap.String("schema", string(dbByte)))
		d.logger.Error("failed to deploy database", zap.Error(err))
		return nil, fmt.Errorf("failed to deploy databse: %w", err)
	}

	d.logger.Debug("deployed database", zap.Any("Resp", out.Result))
	txHash, err = parseRespTxHash(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) TxSuccess(_ context.Context, txHash []byte) error {
	cmd := d.newKwilCliCmd("utils", "query-tx", hex.EncodeToString(txHash))
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to query tx", zap.Error(err))
		return fmt.Errorf("failed to query tx: %w", err)
	}

	d.logger.Debug("tx query result", zap.Any("result", out.Result))

	err = parseRespTxQuery(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx query", zap.Error(err))
		return fmt.Errorf("failed to parse tx query: %w", err)
	}
	return nil
}

func (d *KwilCliDriver) DropDatabase(_ context.Context, dbName string) (txHash []byte, err error) {
	cmd := d.newKwilCliCmd("database", "drop", dbName)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to drop database", zap.Error(err))
		return nil, fmt.Errorf("failed to drop database: %w", err)
	}

	d.logger.Debug("dropped database", zap.Any("Resp", out.Result))
	txHash, err = parseRespTxHash(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) getSchema(dbid string) (*transactions.Schema, error) {
	cmd := d.newKwilCliCmd("database", "read-schema", "--dbid", dbid)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to drop database", zap.Error(err))
		return nil, fmt.Errorf("failed to drop database: %w", err)
	}

	schema, err := parseRespGetSchema(out.Result)
	if err != nil {
		d.logger.Error("failed to parse read schema response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse read schema response: %w", err)
	}
	return schema, nil
}

// prepareCliActionParams returns the named action args for the given action name, in
// the format of `name:value`
func (d *KwilCliDriver) prepareCliActionParams(dbid string, actionName string, actionInputs []any) ([]string, error) {
	schema, err := d.getSchema(dbid)
	if err != nil {
		d.logger.Error("failed to get schema", zap.Error(err))
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
		d.logger.Error("invalid number of inputs", zap.Int("expected", len(action.Inputs)), zap.Int("got", len(actionInputs)))
		return nil, fmt.Errorf("invalid number of inputs, expected %d, got %d", len(action.Inputs), len(actionInputs))
	}

	args := []string{}
	for i, input := range action.Inputs {
		args = append(args, fmt.Sprintf("%s:%v", input, actionInputs[i]))
	}
	return args, nil
}

func (d *KwilCliDriver) ExecuteAction(_ context.Context, dbid string, action string, inputs ...[]any) ([]byte, error) {
	// NOTE: kwil-cli does not support batched inputs
	actionInputs, err := d.prepareCliActionParams(dbid, action, inputs[0])
	if err != nil {
		d.logger.Error("failed to get action params", zap.Error(err))
		return nil, err
	}

	args := []string{"database", "execute", "--dbid", dbid, "--action", action}
	args = append(args, actionInputs...)

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to execute action", zap.Error(err))
		return nil, fmt.Errorf("failed to execute action: %w", err)
	}

	d.logger.Debug("deployed database", zap.Any("Resp", out.Result))
	txHash, err := parseRespTxHash(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) QueryDatabase(_ context.Context, dbid, query string) (*client.Records, error) {
	cmd := d.newKwilCliCmd("database", "query", "--dbid", dbid, query)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to query database", zap.Error(err))
		return nil, fmt.Errorf("failed to d: %w", err)
	}

	d.logger.Debug("query result", zap.Any("result", out.Result))
	records, err := parseRespQueryDb(out.Result)
	if err != nil {
		d.logger.Error("failed to parse query result", zap.Error(err))
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}
	return records, nil
}

func (d *KwilCliDriver) Call(_ context.Context, dbid, action string, inputs []any, withSignature bool) (*client.Records, error) {
	// NOTE: kwil-cli does not support batched inputs
	actionInputs, err := d.prepareCliActionParams(dbid, action, inputs)
	if err != nil {
		d.logger.Error("failed to prepare action params", zap.Error(err))
		return nil, err
	}

	args := []string{"database", "call", "--dbid", dbid, "--action", action}
	args = append(args, actionInputs...)

	if withSignature {
		args = append(args, "--authenticate")
	}

	cmd := d.newKwilCliCmd(args...)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to call action", zap.Error(err))
		return nil, fmt.Errorf("failed to call action: %w", err)
	}

	d.logger.Debug("call result", zap.Any("result", out.Result))
	return parseRespQueryDb(out.Result)
}

func (d *KwilCliDriver) ApproveNode(_ context.Context, joinerPubKey []byte) error {
	cmd := d.newKwilCliCmd("validator", "approve", hex.EncodeToString(joinerPubKey))
	_, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to approve node", zap.Error(err))
		return fmt.Errorf("failed to approve node: %w", err)
	}

	return nil
}

func (d *KwilCliDriver) ValidatorNodeApprove(_ context.Context, joinerPubKey []byte) ([]byte, error) {
	cmd := d.newKwilAdminCmd("validators", "approve", hex.EncodeToString(joinerPubKey), d.privKey)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to approve validator", zap.Error(err))
		return nil, fmt.Errorf("failed to approve validators: %w", err)
	}

	d.logger.Debug("validator approve", zap.Any("Resp", out.Result))
	txHash, err := parseRespTxHash(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) ValidatorNodeRemove(ctx context.Context, target []byte) ([]byte, error) {
	cmd := d.newKwilAdminCmd("validators", "remove", hex.EncodeToString(target), d.privKey)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to approve validators: %w", err)
	}

	d.logger.Debug("validator target", zap.Any("Resp", out.Result))
	txHash, err := parseRespTxHash(out.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) ValidatorNodeJoin(_ context.Context) ([]byte, error) {
	cmd := d.newKwilAdminCmd("validators", "join", d.privKey)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to join as validator", zap.Error(err))
		return nil, fmt.Errorf("failed to joni as validator: %w", err)
	}

	d.logger.Debug("validator join", zap.Any("Resp", out.Result))
	txHash, err := parseRespTxHash(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) ValidatorNodeLeave(_ context.Context) ([]byte, error) {
	cmd := d.newKwilAdminCmd("validators", "leave", d.privKey)
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to leave as validator", zap.Error(err))
		return nil, fmt.Errorf("failed to leave as validator: %w", err)
	}

	d.logger.Debug("validator leave", zap.Any("Resp", out.Result))
	txHash, err := parseRespTxHash(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return txHash, nil
}

func (d *KwilCliDriver) ValidatorJoinStatus(_ context.Context, pubKey []byte) (*types.JoinRequest, error) {
	cmd := d.newKwilAdminCmd("validators", "join-status", hex.EncodeToString(pubKey))
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to query validator join status", zap.Error(err))
		return nil, fmt.Errorf("failed to query validator join status: %w", err)
	}

	d.logger.Debug("validator join status", zap.Any("Resp", out.Result))
	joinReq, err := parseRespValJoinRequest(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return joinReq, nil
}

func (d *KwilCliDriver) ValidatorsList(_ context.Context) ([]*types.Validator, error) {
	cmd := d.newKwilAdminCmd("validators", "list")
	out, err := mustRun(cmd, d.logger)
	if err != nil {
		d.logger.Error("failed to list validators", zap.Error(err))
		return nil, fmt.Errorf("failed to list validators: %w", err)
	}

	d.logger.Debug("validator list", zap.Any("Resp", out.Result))
	valSets, err := parseRespValSets(out.Result)
	if err != nil {
		d.logger.Error("failed to parse tx hash", zap.Error(err))
		return nil, fmt.Errorf("failed to parse tx hash: %w", err)
	}

	return valSets, nil
}

///////// helper functions

// mustRun runs the give command, and parse stdout
func mustRun(cmd *exec.Cmd, logger log.Logger) (*cliResponse, error) {
	cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout
	//// here we ignore the stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	output := out.Bytes()
	logger.Debug("cmd output", zap.String("output", string(output)))

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
	Result any    `json:"result"`
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
func parseRespTxQuery(data any) error {
	bts, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal tx query resp: %w", err)
	}

	var resp respTxQuery
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tx query: %w", err)
	}

	if resp.TxResult.Code != 0 {
		return fmt.Errorf("tx failed: %s", resp.TxResult.Log)
	}

	// NOTE: this should not be considered a failure, should retry
	if resp.Height < 0 {
		return ErrTxNotConfirmed
	}

	return nil
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

// respValJoinRequest is customized json format for respValJoinStatus
// NOTE: this is exactly the same as the one in cmd/kwil-admin/message.go
type respValJoinRequest struct {
	Candidate string `json:"candidate"`
	Power     int64  `json:"power"`
	Board     []string
	Approved  []bool
}

// parseRespValJoinRequest parses the validator join request response(json) from the cli response
// NOTE: this could be defined as a `encoding.TextUnmarshaler` interface in `cmd/kwil-cli`
// if we expose the type from `cmd/kwil-cli`
func parseRespValJoinRequest(data any) (*types.JoinRequest, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal val join request resp: %w", err)
	}

	var resp respValJoinRequest
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal val join request: %w", err)
	}

	candidateBts, err := hex.DecodeString(resp.Candidate)
	if err != nil {
		return nil, fmt.Errorf("failed to decode candidate: %w", err)
	}

	board := make([][]byte, len(resp.Board))
	for i := range resp.Board {
		board[i], err = hex.DecodeString(resp.Board[i])
		if err != nil {
			return nil, fmt.Errorf("failed to decode board: %w", err)
		}
	}

	return &types.JoinRequest{
		Candidate: candidateBts,
		Power:     resp.Power,
		Board:     board,
		Approved:  resp.Approved,
	}, nil
}

// respValInfo represents the validator info response(json) from the cli response
// NOTE: this is exactly the same as the one in cmd/kwil-admin/message.go
type respValInfo struct {
	PubKey string `json:"pubkey"`
	Power  int64  `json:"power"`
}

func parseRespValSets(data any) ([]*types.Validator, error) {
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal val sets resp: %w", err)
	}

	var resp []respValInfo
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal val sets: %w", err)
	}

	vals := make([]*types.Validator, len(resp))
	for i := range resp {
		pubKey, err := hex.DecodeString(resp[i].PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode pubkey: %w", err)
		}
		vals[i] = &types.Validator{
			PubKey: pubKey,
			Power:  resp[i].Power,
		}
	}

	return vals, nil
}
