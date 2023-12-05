// Package client contains the client for interacting with the Kwil public API.
// It's supposed to be used as go-sdk for Kwil, currently used by the Kwil CLI.
package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	"github.com/kwilteam/kwil-db/core/rpc/client/user/http"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"go.uber.org/zap"
	grpcCodes "google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

// Client is a Kwil client that can interact with the main public Kwil RPC.
type Client struct {
	txClient user.TxSvcClient
	Signer   auth.Signer
	logger   log.Logger
	// chainID is used when creating transactions as replay protection since the
	// signatures will only be valid on this network.
	chainID string

	noWarnings bool // silence warning logs
}

// NewClient creates a Kwil client. It will dial the remote host via HTTP, and
// verify the chain ID of the remote host against the chain ID passed in.
func NewClient(ctx context.Context, target string, options *ClientOptions) (c *Client, err error) {
	parsedUrl, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	httpClient := http.NewClient(parsedUrl)

	clt, err := WrapClient(ctx, httpClient, options)
	if err != nil {
		return nil, fmt.Errorf("wrap client: %w", err)
	}

	clt.logger = *clt.logger.Named("client").With(zap.String("host", target))

	return clt, nil
}

// WrapClient wraps an rpc client with a Kwil client.
// It provides a way to use a custom rpc client with the Kwil client.
// Unless a custom rpc client is needed, use Dial instead.
func WrapClient(ctx context.Context, client user.TxSvcClient, options *ClientOptions) (*Client, error) {
	clientOptions := DefaultOptions()
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
		return nil, fmt.Errorf("ping: %w", err)
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

// ChainInfo get the current blockchain information like chain ID and best block
// height/hash.
func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	return c.txClient.ChainInfo(ctx)
}

// GetSchema gets a schema by dbid.
func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	ds, err := c.txClient.GetSchema(ctx, dbid)
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
	return c.txClient.Broadcast(ctx, tx)
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

	res, err := c.txClient.Broadcast(ctx, tx)
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

	return c.txClient.Broadcast(ctx, tx)
}

// CallAction call an action. It returns the result records.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs []any) (*Records, error) {
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

	if c.Signer != nil {
		msg.AuthType = c.Signer.AuthType()
		msg.Sender = c.Signer.Identity()
	}

	res, err := c.txClient.Call(ctx, msg)
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
	res, err := c.txClient.Query(ctx, dbid, query)
	if err != nil {
		return nil, err
	}

	return NewRecordsFromMaps(res), nil
}

func (c *Client) ListDatabases(ctx context.Context, owner []byte) ([]*types.DatasetIdentifier, error) {
	return c.txClient.ListDatabases(ctx, owner)
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.txClient.Ping(ctx)
}

func (c *Client) GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error) {
	return c.txClient.GetAccount(ctx, pubKey, status)
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

func (c *Client) ChainID() string {
	return c.chainID
}
