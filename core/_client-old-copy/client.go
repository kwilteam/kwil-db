// Package client contains the client for interacting with the Kwil public API.
// It's supposed to be used as go-sdk for Kwil, currently used by the Kwil CLI.

package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	rpcClient "github.com/kwilteam/kwil-db/core/rpc/client"
	httpRPC "github.com/kwilteam/kwil-db/core/rpc/client/user/http"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpcCodes "google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

// RPCClient defines the methods that the transport layer should implement.
type RPCClient interface {
	Close() error
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	Call(ctx context.Context, req *transactions.CallMessage, opts ...rpcClient.ActionCallOption) ([]map[string]any, error)
	TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error)
	GetTarget() string
	GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	ListDatabases(ctx context.Context, acctID []byte) ([]string, error)
	GetAccount(ctx context.Context, acctID []byte, status types.AccountStatus) (*types.Account, error)
	Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error)
	Ping(ctx context.Context) (string, error)
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
	ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error)
	CurrentValidators(ctx context.Context) ([]*types.Validator, error)
	VerifySignature(ctx context.Context, sender []byte, signature *auth.Signature, message []byte) error
}

var (
	ErrNotFound = errors.New("not found")
)

// Client wraps the methods to interact with the Kwil public API.
// All the transport level details are encapsulated in the rpc.
type Client struct {
	rpc    RPCClient
	Signer auth.Signer
	logger log.Logger
	// chainID is used when creating transactions as replay protection since the
	// signatures will only be valid on this network.
	chainID     string
	target      string
	cookie      *http.Cookie // current session cookie
	tlsCertFile string       // the tls cert file path

	noWarnings bool // silence warning logs
}

// Dial creates a Kwil client. It will dial the remote host via HTTP, and
// verify the chain ID of the remote host against the chain ID passed in.
func Dial(ctx context.Context, target string, opts ...Option) (c *Client, err error) {
	c = &Client{
		logger: log.NewNoOp(), // by default, we do not want to force client to log anything
		target: target,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.rpc == nil {
		hc, err := httpRPC.Dial(target)
		// NOTE: target will be ignored if WithRPCClient is passed
		if err != nil {
			return nil, err
		}
		c.rpc = hc
	}

	zapFields := []zapcore.Field{
		zap.String("host", c.rpc.GetTarget()),
	}

	c.logger = *c.logger.Named("client").With(zapFields...)

	chainInfo, err := c.ChainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	chainID := chainInfo.ChainID
	if c.chainID == "" {
		if !c.noWarnings {
			c.logger.Warn("chain ID not set, trusting chain ID from remote host!", zap.String("chainID", chainID))
		}
		c.chainID = chainID
	} else if c.chainID != chainID {
		c.Close()
		return nil, fmt.Errorf("remote host chain ID %q != client configured %q", chainID, c.chainID)
	}

	return c, nil
}

func (c *Client) Close() error {
	return c.rpc.Close()
}

// ChainInfo get the current blockchain information like chain ID and best block
// height/hash.
func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	return c.rpc.ChainInfo(ctx)
}

// GetSchema gets a schema by dbid.
func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	ds, err := c.rpc.GetSchema(ctx, dbid)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

// DeployDatabase deploys a schema
func (c *Client) DeployDatabase(ctx context.Context, payload *transactions.Schema, opts ...TxOpt) (transactions.TxHash, error) {
	tx, err := c.newTx(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)))

	return c.rpc.Broadcast(ctx, tx)
}

// DropDatabase drops a database by name, using the configured signer to derive
// the DB ID.
func (c *Client) DropDatabase(ctx context.Context, name string, opts ...TxOpt) (transactions.TxHash, error) {
	dbid := utils.GenerateDBID(name, c.Signer.Identity())
	return c.DropDatabaseID(ctx, dbid, opts...)
}

// DropDatabaseID drops a database by ID.
func (c *Client) DropDatabaseID(ctx context.Context, dbid string, opts ...TxOpt) (transactions.TxHash, error) {
	identifier := &transactions.DropSchema{
		DBID: dbid,
	}

	tx, err := c.newTx(ctx, identifier, opts...)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)))

	res, err := c.rpc.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ExecuteAction executes an action.
// It returns the receipt, as well as outputs which is the decoded body of the receipt.
// It can take any number of inputs, and if multiple tuples of inputs are passed, it will execute them transactionally.
func (c *Client) ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...TxOpt) (transactions.TxHash, error) {
	stringTuples, err := convertTuples(tuples)
	if err != nil {
		return nil, err
	}

	executionBody := &transactions.ActionExecution{
		Action:    action,
		DBID:      dbid,
		Arguments: stringTuples,
	}

	tx, err := c.newTx(ctx, executionBody, opts...)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("execute action",
		zap.String("signature_type", tx.Signature.Type),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)))

	return c.rpc.Broadcast(ctx, tx)
}

// CallAction call an action, if auxiliary `mustsign` is set, need to sign the action payload. It returns the records.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs []any, opts ...CallOpt) (*Records, error) {
	callOpts := &callOptions{} // TODO: remove

	for _, opt := range opts {
		opt(callOpts)
	}

	stringInputs, err := convertTuple(inputs)
	if err != nil {
		return nil, err
	}

	payload := &transactions.ActionCall{
		DBID:      dbid,
		Action:    action,
		Arguments: stringInputs,
	}

	msg, err := transactions.CreateCallMessage(payload)
	if err != nil {
		return nil, fmt.Errorf("create signed message: %w", err)
	}

	msg.AuthType = c.Signer.AuthType()
	msg.Sender = c.Signer.Identity()

	// NOTE: only HTTP RPC supports authn cookie
	var actCallOpts []rpcClient.ActionCallOption
	if c.cookie != nil {
		actCallOpts = append(actCallOpts, rpcClient.WithAuthCookie(c.cookie))
	}
	res, err := c.rpc.Call(ctx, msg, actCallOpts...)
	if err != nil {
		return nil, fmt.Errorf("call action: %w", err)
	}

	return NewRecordsFromMaps(res), nil
}

func DecodeOutputs(bts []byte) ([]map[string]any, error) {
	if len(bts) == 0 {
		return []map[string]any{}, nil
	}

	var outputs []map[string]any
	err := json.Unmarshal(bts, &outputs)
	if err != nil {
		return nil, err
	}

	return outputs, nil
}

// Query executes a query
func (c *Client) Query(ctx context.Context, dbid string, query string) (*Records, error) {
	res, err := c.rpc.Query(ctx, dbid, query)
	if err != nil {
		return nil, err
	}

	return NewRecordsFromMaps(res), nil
}

func (c *Client) ListDatabases(ctx context.Context, owner []byte) ([]string, error) {
	return c.rpc.ListDatabases(ctx, owner)
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.rpc.Ping(ctx)
}

func (c *Client) GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error) {
	return c.rpc.GetAccount(ctx, pubKey, status)
}

func (c *Client) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error) {
	res, err := c.rpc.ValidatorJoinStatus(ctx, pubKey)
	if err != nil {
		if stat, ok := grpcStatus.FromError(err); ok {
			if stat.Code() == grpcCodes.NotFound {
				return nil, ErrNotFound
			}
		}
		return nil, err
	}
	return res, nil
}

func (c *Client) CurrentValidators(ctx context.Context) ([]*types.Validator, error) {
	return c.rpc.CurrentValidators(ctx)
}

func (c *Client) ApproveValidator(ctx context.Context, joiner []byte, opts ...TxOpt) ([]byte, error) {
	_, err := crypto.Ed25519PublicKeyFromBytes(joiner)
	if err != nil {
		return nil, fmt.Errorf("invalid candidate validator public key: %w", err)
	}
	payload := &transactions.ValidatorApprove{
		Candidate: joiner,
	}
	tx, err := c.newTx(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	hash, err := c.rpc.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// RemoveValidator makes a transaction proposing to remove a validator. This is
// only useful if the Client's signing key is a current validator key.
func (c *Client) RemoveValidator(ctx context.Context, target []byte, opts ...TxOpt) ([]byte, error) {
	_, err := crypto.Ed25519PublicKeyFromBytes(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target validator public key: %w", err)
	}
	payload := &transactions.ValidatorRemove{
		Validator: target,
	}
	tx, err := c.newTx(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	return c.rpc.Broadcast(ctx, tx)
}

func (c *Client) ValidatorJoin(ctx context.Context) ([]byte, error) {
	const power = 1
	return c.validatorUpdate(ctx, power)
}

func (c *Client) ValidatorLeave(ctx context.Context) ([]byte, error) {
	return c.validatorUpdate(ctx, 0)
}

func (c *Client) validatorUpdate(ctx context.Context, power int64, opts ...TxOpt) ([]byte, error) {
	var payload transactions.Payload
	if power <= 0 {
		payload = &transactions.ValidatorLeave{}
	} else {
		payload = &transactions.ValidatorJoin{
			Power: uint64(power),
		}
	}

	tx, err := c.newTx(ctx, payload, opts...)
	if err != nil {
		return nil, err
	}

	hash, err := c.rpc.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// convertTuples converts user passed tuples to strings.
// this is necessary for RLP encoding
func convertTuples(tuples [][]any) ([][]string, error) {
	ins := [][]string{}
	for _, tuple := range tuples {
		stringTuple, err := convertTuple(tuple)
		if err != nil {
			return nil, err
		}
		ins = append(ins, stringTuple)
	}

	return ins, nil
}

// convertTuple converts user passed tuple to strings.
func convertTuple(tuple []any) ([]string, error) {
	stringTuple := []string{}
	for _, val := range tuple {

		stringVal, err := conv.String(val)
		if err != nil {
			return nil, err
		}

		stringTuple = append(stringTuple, stringVal)
	}

	return stringTuple, nil
}

// TxQuery get transaction by hash
func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	res, err := c.rpc.TxQuery(ctx, txHash)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// WaitTx repeatedly queries at a given interval for the status of a transaction
// until it is confirmed (is included in a block).
func (c *Client) WaitTx(ctx context.Context, txHash []byte, interval time.Duration) (*transactions.TcTxQueryResponse, error) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		resp, err := c.TxQuery(ctx, txHash)
		notFound := grpcStatus.Code(err) == grpcCodes.NotFound
		if !notFound {
			if err != nil {
				return nil, err
			}
			if resp.Height > 0 {
				return resp, nil
			}
		} else {
			// NOTE: this log may be removed once we've resolved the issue of
			// transactions not being found immediately after broadcast.
			c.logger.Debug("tx not found")
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// VerifySignature verifies a signature through API.
// An ErrInvalidSignature is returned if the signature is invalid.
func (c *Client) VerifySignature(ctx context.Context, pubKey []byte,
	signature *auth.Signature, message []byte) error {
	return c.rpc.VerifySignature(ctx, pubKey, signature, message)
}

// GatewayAuthenticate authenticates the client with the gateway(KGW) provider.
// It returns a cookie so caller can choose a way to persist it if wanted.
// It also sets the cookie in the client, so that it can be used for subsequent requests.
// 'signMessage' is a function to present the message to be signed to
// the user, and returns error if the user declines to sign the message.
func (c *Client) GatewayAuthenticate(ctx context.Context,
	signMessage func(message string) error) (*http.Cookie, error) {
	// KGW auth is not part of Kwil API, we use a standard http client
	// this client is to reuse connection
	hc := httpRPC.DefaultHTTPClient()

	authURI, err := url.JoinPath(c.target, kgwAuthEndpoint)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	authParam, err := requestGatewayAuthParameter(hc, authURI)
	if err != nil {
		return nil, fmt.Errorf("request for authentication: %w", err)
	}

	msg := composeGatewayAuthMessage(authParam, c.target, authURI, kgwAuthVersion, c.chainID)
	decline := signMessage(msg)
	if decline != nil {
		return nil, decline
	}

	sig, err := c.Signer.Sign([]byte(msg))
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}

	cookie, err := requestGatewayAuthCookie(hc, c.target, authParam.Nonce,
		c.Signer.Identity(), sig)
	if err != nil {
		return nil, fmt.Errorf("request for token: %w", err)
	}

	// set cookie for current session
	c.cookie = cookie
	return cookie, nil
}

// SetSigner sets a new signer and chain ID for the client.
// If either is left nil, the client will keep the old value.
func (c *Client) SetSigner(signer auth.Signer, chainID string) {
	if signer != nil {
		c.Signer = signer
	}
	if chainID != "" {
		c.chainID = chainID
	}
}
