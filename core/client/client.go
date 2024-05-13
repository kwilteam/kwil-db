// Package client defines client for interacting with the Kwil provider.
// It's supposed to be used as go-sdk for Kwil, currently used by the Kwil CLI.
package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	userClient "github.com/kwilteam/kwil-db/core/rpc/client/user/jsonrpc"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"go.uber.org/zap"
)

// Client is a client that interacts with a public Kwil provider.
type Client struct {
	txClient user.TxSvcClient
	Signer   auth.Signer
	logger   log.Logger
	// chainID is used when creating transactions as replay protection since the
	// signatures will only be valid on this network.
	chainID string

	noWarnings bool // silence warning logs
}

var _ clientType.Client = (*Client)(nil)

// NewClient creates a Kwil client. The target should be a URL (for an
// http.Client). It by default communicates with target via HTTP; chain ID of the
// remote host will be verified against the chain ID passed in.
func NewClient(ctx context.Context, target string, options *clientType.Options) (c *Client, err error) {
	// OPTION A: Target is a base URL, and the jsonrpc client appends the path
	// for the API version it speaks (e.g. /rpc/v1). The json rpc client knows
	// v1 methods and types, and speaks to the /rpc/v1 handler.
	parsedURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	// OPTION B: The jsonrpc client uses the URL unmodified. No path is
	// appended. Caller must use /rpc/v1 or whatever directs to the v1 handler.
	// If a hostport is provided, construct a URL with the default scheme
	// (http://) and path (/rpc/v1). Preferred use is with a URL.
	//
	// _, _, err = net.SplitHostPort(target)
	// if err == nil {
	// 	target = "http://" + target + "/rpc/v1"
	// } else {
	// 	_, err = url.Parse(target)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("parse url: %w", err)
	// 	}
	// }

	jsonrpcClientOpts := []rpcclient.RPCClientOpts{}
	if options != nil && options.Logger.L != nil {
		jsonrpcClientOpts = append(jsonrpcClientOpts, rpcclient.WithLogger(options.Logger))
	}
	client := userClient.NewClient(parsedURL, jsonrpcClientOpts...)

	clt, err := WrapClient(ctx, client, options)
	if err != nil {
		return nil, fmt.Errorf("wrap client: %w", err)
	}

	clt.logger = *clt.logger.Named("client").With(zap.String("host", target))

	return clt, nil
}

// WrapClient wraps a TxSvcClient with a Kwil client.
// It provides a way to use a custom rpc client with the Kwil client.
func WrapClient(ctx context.Context, client user.TxSvcClient, options *clientType.Options) (*Client, error) {
	clientOptions := clientType.DefaultOptions()
	clientOptions.Apply(options)

	c := &Client{
		txClient:   client,
		Signer:     clientOptions.Signer,
		logger:     clientOptions.Logger,
		chainID:    clientOptions.ChainID,
		noWarnings: clientOptions.Silence,
	}

	chainInfo, err := c.ChainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain_info: %w", err)
	}
	chainID := chainInfo.ChainID
	if c.chainID == "" {
		if !c.noWarnings {
			c.logger.Warn("chain ID not set, trusting chain ID from remote host!", zap.String("chainID", chainID))
		}
		c.chainID = chainID
	} else if c.chainID != chainID {
		return nil, fmt.Errorf("remote host chain ID %q != client configured %q", chainID, c.chainID)
	}

	return c, nil
}

func syncBcastFlag(syncBcast bool) rpcclient.BroadcastWait {
	syncFlag := rpcclient.BroadcastWaitSync
	if syncBcast { // the bool really means wait for commit in cometbft terms
		syncFlag = rpcclient.BroadcastWaitCommit
	}
	return syncFlag
}

// Transfer transfers balance to a given address.
func (c *Client) Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	// Get account balance to ensure we can afford the transfer, and use the
	// nonce to avoid a second GetAccount in newTx.
	acct, err := c.txClient.GetAccount(ctx, c.Signer.Identity(), types.AccountStatusPending)
	if err != nil {
		return nil, err
	}
	nonceOpt := clientType.WithNonce(acct.Nonce + 1)
	opts = append([]clientType.TxOpt{nonceOpt}, opts...) // prepend in case caller specified a nonce
	txOpts := clientType.GetTxOpts(opts)

	trans := &transactions.Transfer{
		To:     to,
		Amount: amount.String(),
	}
	tx, err := c.newTx(ctx, trans, txOpts)
	if err != nil {
		return nil, err
	}

	totalSpend := big.NewInt(0).Add(tx.Body.Fee, amount)
	if totalSpend.Cmp(acct.Balance) > 0 {
		return nil, fmt.Errorf("send amount plus fees (%v) larger than balance (%v)", totalSpend, acct.Balance)
	}

	c.logger.Debug("transfer", zap.String("to", hex.EncodeToString(to)),
		zap.String("amount", amount.String()))

	return c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
}

// ChainInfo get the current blockchain information like chain ID and best block
// height/hash.
func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	return c.txClient.ChainInfo(ctx)
}

// GetSchema gets a schema by dbid.
func (c *Client) GetSchema(ctx context.Context, dbid string) (*types.Schema, error) {
	ds, err := c.txClient.GetSchema(ctx, dbid)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

// DeployDatabase deploys a database.
func (c *Client) DeployDatabase(ctx context.Context, payload *types.Schema, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	txOpts := clientType.GetTxOpts(opts)
	s2 := &transactions.Schema{}
	s2.FromTypes(payload)
	tx, err := c.newTx(ctx, s2, txOpts)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)),
		zap.String("fee", tx.Body.Fee.String()), zap.Int64("nonce", int64(tx.Body.Nonce)))
	return c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
}

// DropDatabase drops a database by name, using the configured signer to derive
// the DB ID.
func (c *Client) DropDatabase(ctx context.Context, name string, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	dbid := utils.GenerateDBID(name, c.Signer.Identity())
	return c.DropDatabaseID(ctx, dbid, opts...)
}

// DropDatabaseID drops a database by ID.
func (c *Client) DropDatabaseID(ctx context.Context, dbid string, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	identifier := &transactions.DropSchema{
		DBID: dbid,
	}

	txOpts := clientType.GetTxOpts(opts)
	tx, err := c.newTx(ctx, identifier, txOpts)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)),
		zap.String("fee", tx.Body.Fee.String()), zap.Int64("nonce", int64(tx.Body.Nonce)))

	res, err := c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DEPRECATED: Use Execute instead.
func (c *Client) ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	return c.Execute(ctx, dbid, action, tuples, opts...)
}

// Execute executes a procedure or action.
// It returns the receipt, as well as outputs which is the decoded body of the receipt.
// It can take any number of inputs, and if multiple tuples of inputs are passed,
// it will execute them in the same transaction.
func (c *Client) Execute(ctx context.Context, dbid string, procedure string, tuples [][]any, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	stringTuples, isNil, err := convertTuples(tuples)
	if err != nil {
		return nil, err
	}

	executionBody := &transactions.ActionExecution{
		Action:    procedure,
		DBID:      dbid,
		Arguments: stringTuples,
		NilArg:    isNil,
	}

	txOpts := clientType.GetTxOpts(opts)
	tx, err := c.newTx(ctx, executionBody, txOpts)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("execute action",
		zap.String("DBID", dbid), zap.String("action", procedure),
		zap.String("signature_type", tx.Signature.Type),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)),
		zap.String("fee", tx.Body.Fee.String()), zap.Int64("nonce", int64(tx.Body.Nonce)))

	return c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
}

// DEPRECATED: Use Call instead.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs []any) (*clientType.Records, error) {
	return c.Call(ctx, dbid, action, inputs)
}

// Call calls a procedure or action. It returns the result records.
func (c *Client) Call(ctx context.Context, dbid string, procedure string, inputs []any) (*clientType.Records, error) {
	stringInputs, isNil, err := convertTuple(inputs)
	if err != nil {
		return nil, err
	}

	payload := &transactions.ActionCall{
		DBID:      dbid,
		Action:    procedure,
		Arguments: stringInputs,
		NilArg:    isNil,
	}

	msg, err := transactions.CreateCallMessage(payload)
	if err != nil {
		return nil, fmt.Errorf("create signed message: %w", err)
	}

	if c.Signer != nil {
		msg.AuthType = c.Signer.AuthType()
		msg.Sender = c.Signer.Identity()
	}

	res, err := c.txClient.Call(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("call action: %w", err)
	}

	return clientType.NewRecordsFromMaps(res), nil
}

// Query executes a query.
func (c *Client) Query(ctx context.Context, dbid string, query string) (*clientType.Records, error) {
	res, err := c.txClient.Query(ctx, dbid, query)
	if err != nil {
		return nil, err
	}

	return clientType.NewRecordsFromMaps(res), nil
}

// ListDatabases lists databases belonging to an owner.
// If no owner is passed, it will list all databases.
func (c *Client) ListDatabases(ctx context.Context, owner []byte) ([]*types.DatasetIdentifier, error) {
	return c.txClient.ListDatabases(ctx, owner)
}

// Ping pings the remote host.
func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.txClient.Ping(ctx)
}

// GetAccount gets account info by account ID.
// If status is AccountStatusPending, it will include the pending info.
func (c *Client) GetAccount(ctx context.Context, acctID []byte, status types.AccountStatus) (*types.Account, error) {
	return c.txClient.GetAccount(ctx, acctID, status)
}

// convertTuples converts user passed tuples to strings.
// this is necessary for RLP encoding
func convertTuples(tuples [][]any) ([][]string, [][]bool, error) {
	ins := make([][]string, 0, len(tuples))
	nils := make([][]bool, 0, len(tuples))
	for _, tuple := range tuples {
		stringTuple, isNil, err := convertTuple(tuple)
		if err != nil {
			return nil, nil, err
		}
		ins = append(ins, stringTuple)
		nils = append(nils, isNil)
	}

	return ins, nils, nil
}

// convertTuple converts user passed tuple to strings.
func convertTuple(tuple []any) ([]string, []bool, error) {
	stringTuple := make([]string, 0, len(tuple))
	isNil := make([]bool, 0, len(tuple))
	for _, val := range tuple {
		if val == nil {
			stringTuple = append(stringTuple, "")
			isNil = append(isNil, true)
			continue
		}

		// conv.String would make it "<null>", which could very well be an intended string
		stringVal, err := conv.String(val)
		if err != nil {
			return nil, nil, err
		}

		stringTuple = append(stringTuple, stringVal)
		isNil = append(isNil, false)
	}

	return stringTuple, isNil, nil
}

// TxQuery get transaction by hash.
func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	res, err := c.txClient.TxQuery(ctx, txHash)
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
		if err != nil {
			// Only error out if it's something other than not found.
			if !errors.Is(err, rpcclient.ErrNotFound) {
				return nil, err
			} // else not found, try again next time
		} else if resp.Height > 0 {
			return resp, nil
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// ChainID returns the chain ID of the remote host.
func (c *Client) ChainID() string {
	return c.chainID
}
