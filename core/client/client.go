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

	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	userClient "github.com/kwilteam/kwil-db/core/rpc/client/user/jsonrpc"
	"github.com/kwilteam/kwil-db/core/types"
)

// Client is a client that interacts with a public Kwil provider.
type Client struct {
	txClient user.TxSvcClient
	Signer   auth.Signer
	logger   log.Logger
	// chainID is used when creating transactions as replay protection since the
	// signatures will only be valid on this network.
	chainID string

	// skipVerifyChainID skip checking chain ID against remote node's chain ID.
	// This is only effective when chainID is set.
	skipVerifyChainID bool

	// skipHealthcheck will skip check remote nodes' health status.
	skipHealthcheck bool

	noWarnings bool // silence warning logs

	authCallRPC bool
}

// SvcClient is a trapdoor to access the underlying
// core/rpc/client/user.TxSvcClient. Most applications will only use the methods
// of Client.
func (c *Client) SvcClient() user.TxSvcClient {
	return c.txClient
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
	if options != nil && options.Logger != nil {
		jsonrpcClientOpts = append(jsonrpcClientOpts, rpcclient.WithLogger(options.Logger))
	}
	client := userClient.NewClient(parsedURL, jsonrpcClientOpts...)

	clt, err := WrapClient(ctx, client, options)
	if err != nil {
		return nil, fmt.Errorf("wrap client: %w", err)
	}

	clt.logger = clt.logger.New("client")

	return clt, nil
}

// WrapClient wraps a TxSvcClient with a Kwil client.
// It provides a way to use a custom rpc client with the Kwil client.
func WrapClient(ctx context.Context, client user.TxSvcClient, options *clientType.Options) (*Client, error) {
	clientOptions := clientType.DefaultOptions()
	clientOptions.Apply(options)

	c := &Client{
		txClient:          client,
		Signer:            clientOptions.Signer,
		logger:            clientOptions.Logger,
		chainID:           clientOptions.ChainID,
		noWarnings:        clientOptions.Silence,
		skipVerifyChainID: clientOptions.SkipVerifyChainID,
		skipHealthcheck:   clientOptions.SkipHealthcheck,
	}

	var remoteChainID string

	if c.skipHealthcheck {
		health, err := c.Health(ctx)
		// NOTE: we ignore all errors from c.Health call since we ignore health check
		if err == nil {
			// this is v09 API, we just take the result.
			c.authCallRPC = health.Mode == types.ModePrivate
			remoteChainID = health.ChainID

			// NOTE: since original health check only log, why not ?
			if health.Healthy {
				c.logger.Warnf("node reports that it is not healthy: %v", health)
			}
		} else {
			// fall back to v08 API
			chainInfo, err := c.ChainInfo(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve the node's chain info: %w", err)
			}

			remoteChainID = chainInfo.ChainID
		}
	} else {
		health, err := c.Health(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve the node's health: %w", err)
		}

		if health.Healthy {
			c.logger.Warnf("node reports that it is not healthy: %v", health)
		}

		c.authCallRPC = health.Mode == types.ModePrivate
		remoteChainID = health.ChainID
	}

	if c.chainID == "" { // always use chain ID from remote host
		if !c.noWarnings {
			c.logger.Warn("chain ID not set, trusting chain ID from remote host!",
				"chainID", remoteChainID)
		}

		c.chainID = remoteChainID
	} else {
		if c.skipVerifyChainID {
			if !c.noWarnings {
				c.logger.Warn("chain ID is set, skip check against remote chain ID", "chainID", c.chainID)
			}
		} else if remoteChainID != c.chainID {
			return nil, fmt.Errorf("remote host chain ID %q != client configured %q", remoteChainID, c.chainID)
		}
	}

	return c, nil
}

// PrivateMode returns if it the client has connected to an RPC server that is
// running in "private" mode where call requests require authentication. In
// addition, queries are expected to be denied, and no verbose transaction
// information will returned with a transaction status query.
func (c *Client) PrivateMode() bool {
	return c.authCallRPC
}

func syncBcastFlag(syncBcast bool) rpcclient.BroadcastWait {
	syncFlag := rpcclient.BroadcastWaitSync
	if syncBcast { // the bool really means wait for commit in cometbft terms
		syncFlag = rpcclient.BroadcastWaitCommit
	}
	return syncFlag
}

// Transfer transfers balance to a given address.
func (c *Client) Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...clientType.TxOpt) (types.Hash, error) {
	// Get account balance to ensure we can afford the transfer, and use the
	// nonce to avoid a second GetAccount in newTx.
	acct, err := c.txClient.GetAccount(ctx, c.Signer.Identity(), types.AccountStatusPending)
	if err != nil {
		return types.Hash{}, err
	}
	nonceOpt := clientType.WithNonce(acct.Nonce + 1)
	opts = append([]clientType.TxOpt{nonceOpt}, opts...) // prepend in case caller specified a nonce
	txOpts := clientType.GetTxOpts(opts)

	trans := &types.Transfer{
		To:     to,
		Amount: amount.String(),
	}
	tx, err := c.newTx(ctx, trans, txOpts)
	if err != nil {
		return types.Hash{}, err
	}

	totalSpend := big.NewInt(0).Add(tx.Body.Fee, amount)
	if totalSpend.Cmp(acct.Balance) > 0 {
		return types.Hash{}, fmt.Errorf("send amount plus fees (%v) larger than balance (%v)", totalSpend, acct.Balance)
	}

	c.logger.Debug("transfer", "to", hex.EncodeToString(to),
		"amount", amount.String())

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

/*
// DeployDatabase deploys a database. TODO: remove
func (c *Client) DeployDatabase(ctx context.Context, schema *types.Schema, opts ...clientType.TxOpt) (types.TxHash, error) {
	txOpts := clientType.GetTxOpts(opts)
	tx, err := c.newTx(ctx, schema, txOpts)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		"signature_type", tx.Signature.Type,
		"signature", base64.StdEncoding.EncodeToString(tx.Signature.Data),
		"fee", tx.Body.Fee.String(), "nonce", tx.Body.Nonce)
	return c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
}

// DropDatabase drops a database by name, using the configured signer to derive
// the DB ID. TODO: remove
func (c *Client) DropDatabase(ctx context.Context, name string, opts ...clientType.TxOpt) (types.TxHash, error) {
	dbid := utils.GenerateDBID(name, c.Signer.Identity())
	return c.DropDatabaseID(ctx, dbid, opts...)
}

// DropDatabaseID drops a database by ID. TODO: remove
func (c *Client) DropDatabaseID(ctx context.Context, dbid string, opts ...clientType.TxOpt) (types.TxHash, error) {
	identifier := &types.DropSchema{
		DBID: dbid,
	}

	txOpts := clientType.GetTxOpts(opts)
	tx, err := c.newTx(ctx, identifier, txOpts)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		"signature_type", tx.Signature.Type,
		"signature", base64.StdEncoding.EncodeToString(tx.Signature.Data),
		"fee", tx.Body.Fee.String(), "nonce", tx.Body.Nonce)

	res, err := c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
	if err != nil {
		return nil, err
	}

	return res, nil
}
*/

// Execute executes a procedure or action.
// It returns the receipt, as well as outputs which is the decoded body of the receipt.
// It can take any number of inputs, and if multiple tuples of inputs are passed,
// it will execute them in the same transaction.
func (c *Client) Execute(ctx context.Context, dbid string, procedure string, tuples [][]any, opts ...clientType.TxOpt) (types.Hash, error) {
	encodedTuples := make([][]*types.EncodedValue, len(tuples))
	for i, tuple := range tuples {
		encoded, err := encodeTuple(tuple)
		if err != nil {
			return types.Hash{}, err
		}
		encodedTuples[i] = encoded
	}

	executionBody := &types.ActionExecution{
		Action:    procedure,
		DBID:      dbid,
		Arguments: encodedTuples,
	}

	txOpts := clientType.GetTxOpts(opts)
	tx, err := c.newTx(ctx, executionBody, txOpts)
	if err != nil {
		return types.Hash{}, err
	}

	c.logger.Debug("execute action",
		"DBID", dbid, "action", procedure,
		"signature_type", tx.Signature.Type,
		"signature", base64.StdEncoding.EncodeToString(tx.Signature.Data),
		"fee", tx.Body.Fee.String(), "nonce", tx.Body.Nonce)

	return c.txClient.Broadcast(ctx, tx, syncBcastFlag(txOpts.SyncBcast))
}

// DEPRECATED: Use Call instead.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs []any) (*clientType.Records, error) {
	r, err := c.Call(ctx, dbid, action, inputs)
	if err != nil {
		return nil, err
	}

	return r.Records, nil
}

// Call calls a procedure or action. It returns the result records.
func (c *Client) Call(ctx context.Context, dbid string, procedure string, inputs []any) (*clientType.CallResult, error) {
	encoded, err := encodeTuple(inputs)
	if err != nil {
		return nil, err
	}

	payload := &types.ActionCall{
		DBID:      dbid,
		Action:    procedure,
		Arguments: encoded,
	}

	// If using authenticated call RPCs, request a challenge to include in the
	// signed message text.
	var challenge []byte
	if c.authCallRPC {
		if c.Signer == nil {
			return nil, errors.New("a signer is required with authenticated call RPCs")
		}
		challenge, err = c.challenge(ctx)
		if err != nil {
			return nil, err
		}
	}

	msg, err := types.CreateCallMessage(payload, challenge, c.Signer)
	if err != nil {
		return nil, fmt.Errorf("create signed message: %w", err)
	}

	res, logs, err := c.txClient.Call(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("call action: %w", err)
	}

	return &clientType.CallResult{
		Records: clientType.NewRecordsFromMaps(res),
		Logs:    logs,
	}, nil
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

// encodeTuple encodes a tuple for usage in a transaction.
func encodeTuple(tup []any) ([]*types.EncodedValue, error) {
	encoded := make([]*types.EncodedValue, 0, len(tup))
	for _, val := range tup {
		ev, err := types.EncodeValue(val)
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, ev)
	}

	return encoded, nil
}

// TxQuery get transaction by hash.
func (c *Client) TxQuery(ctx context.Context, txHash types.Hash) (*types.TcTxQueryResponse, error) {
	res, err := c.txClient.TxQuery(ctx, txHash)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// WaitTx repeatedly queries at a given interval for the status of a transaction
// until it is confirmed (is included in a block).
func (c *Client) WaitTx(ctx context.Context, txHash types.Hash, interval time.Duration) (*types.TcTxQueryResponse, error) {
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

// ChainID returns the configured chain ID.
func (c *Client) ChainID() string {
	return c.chainID
}

func (c *Client) ListMigrations(ctx context.Context) ([]*types.Migration, error) {
	return c.txClient.ListMigrations(ctx)
}

func (c *Client) LoadChangeset(ctx context.Context, height int64, index int64) ([]byte, error) {
	return c.txClient.LoadChangeset(ctx, height, index)
}

func (c *Client) ChangesetMetadata(ctx context.Context, height int64) (numChangesets int64, chunkSizes []int64, err error) {
	return c.txClient.ChangesetMetadata(ctx, height)
}

func (c *Client) GenesisState(ctx context.Context) (*types.MigrationMetadata, error) {
	return c.txClient.GenesisState(ctx)
}

func (c *Client) GenesisSnapshotChunk(ctx context.Context, height uint64, chunkIdx uint32) ([]byte, error) {
	return c.txClient.GenesisSnapshotChunk(ctx, height, chunkIdx)
}

func (c *Client) challenge(ctx context.Context) ([]byte, error) {
	return c.txClient.Challenge(ctx)
}

func (c *Client) Health(ctx context.Context) (*types.Health, error) {
	return c.txClient.Health(ctx)
}
