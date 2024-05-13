// Package jsonrpc implements the core/rpc/client/user.TxSvcClient interface
// that is required by core/client.Client.
package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// Client is a JSON-RPC client for the Kwil user service. It use the JSONRPCClient
// from the rpcclient package for the actual JSON-RPC communication, and implements
// the user.TxSvcClient interface.
type Client struct {
	*rpcclient.JSONRPCClient
}

func NewClient(url *url.URL, opts ...rpcclient.RPCClientOpts) *Client {
	return &Client{
		JSONRPCClient: rpcclient.NewJSONRPCClient(url, opts...),
	}
}

var _ user.TxSvcClient = (*Client)(nil)

func (cl *Client) Ping(ctx context.Context) (string, error) {
	cmd := &jsonrpc.PingRequest{
		Message: "ping",
	}
	res := &jsonrpc.PingResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodPing), cmd, res)
	if err != nil {
		return "", err
	}
	return res.Message, nil
}

func (cl *Client) Broadcast(ctx context.Context, tx *transactions.Transaction, sync rpcclient.BroadcastWait) ([]byte, error) {
	cmd := &jsonrpc.BroadcastRequest{
		Tx:   tx,
		Sync: (*jsonrpc.BroadcastSync)(&sync),
	}
	res := &jsonrpc.BroadcastResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodBroadcast), cmd, res)
	if err != nil {
		var jsonRPCErr *jsonrpc.Error
		if errors.As(err, &jsonRPCErr) && jsonRPCErr.Code == jsonrpc.ErrorTxExecFailure && len(jsonRPCErr.Data) > 0 {
			var berr jsonrpc.BroadcastError
			jsonErr := json.Unmarshal(jsonRPCErr.Data, &berr)
			if jsonErr != nil {
				return nil, errors.Join(jsonErr, err)
			}

			switch transactions.TxCode(berr.TxCode) {
			case transactions.CodeWrongChain:
				return nil, errors.Join(transactions.ErrWrongChain, err)
			case transactions.CodeInvalidNonce:
				return nil, errors.Join(transactions.ErrInvalidNonce, err)
			case transactions.CodeInvalidAmount:
				return nil, errors.Join(transactions.ErrInvalidAmount, err)
			case transactions.CodeInsufficientBalance:
				return nil, errors.Join(transactions.ErrInsufficientBalance, err)
			}
		}
		return nil, err
	}
	return res.TxHash, nil
}

func unmarshalMapResults(b []byte) ([]map[string]any, error) {
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.UseNumber()

	// unmashal result
	var result []map[string]any
	err := d.Decode(&result)
	if err != nil {
		return nil, err
	}

	// convert numbers to int64
	for _, record := range result {
		for k, v := range record {
			if num, ok := v.(json.Number); ok {
				i, err := num.Int64()
				if err != nil {
					return nil, err
				}
				record[k] = i
			}
		}
	}

	return result, nil
}

func (cl *Client) Call(ctx context.Context, msg *transactions.CallMessage, opts ...rpcclient.ActionCallOption) ([]map[string]any, error) {
	cmd := &jsonrpc.CallRequest{
		Body:     msg.Body,
		AuthType: msg.AuthType,
		Sender:   msg.Sender,
	}
	res := &jsonrpc.CallResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodCall), cmd, res)
	if err != nil {
		return nil, err
	}
	return unmarshalMapResults(res.Result)
}

func (cl *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	cmd := &jsonrpc.ChainInfoRequest{}
	res := &jsonrpc.ChainInfoResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodChainInfo), cmd, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cl *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	cmd := &jsonrpc.EstimatePriceRequest{
		Tx: tx,
	}
	res := &jsonrpc.EstimatePriceResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodPrice), cmd, res)
	if err != nil {
		return nil, err
	}

	// parse result.Price to big.Int
	price, ok := new(big.Int).SetString(res.Price, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse price to big.Int. received: %s", res.Price)
	}

	return price, nil
}

func (cl *Client) GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error) {
	cmd := &jsonrpc.AccountRequest{
		Identifier: pubKey,
		Status:     &status,
	}
	res := &jsonrpc.AccountResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodAccount), cmd, res)
	if err != nil {
		return nil, err
	}

	// parse result.Balance to big.Int
	balance, ok := new(big.Int).SetString(res.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance to big.Int. received: %s", res.Balance)
	}

	// I'm not sure about nonce yet, could be string could be *big.Int.
	// parsedNonce, err := strconv.ParseInt(res.Account.Nonce, 10, 64)
	// if err != nil {
	// 	return nil, err
	// }

	return &types.Account{
		Identifier: pubKey,
		Balance:    balance,
		Nonce:      res.Nonce,
	}, nil
}

func (cl *Client) GetSchema(ctx context.Context, dbid string) (*types.Schema, error) {
	cmd := &jsonrpc.SchemaRequest{
		DBID: dbid,
	}
	res := &jsonrpc.SchemaResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodSchema), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Schema, nil
}

func (cl *Client) ListDatabases(ctx context.Context, ownerPubKey []byte) ([]*types.DatasetIdentifier, error) {
	cmd := &jsonrpc.ListDatabasesRequest{
		Owner: ownerPubKey,
	}
	res := &jsonrpc.ListDatabasesResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodDatabases), cmd, res)
	if err != nil {
		return nil, err
	}
	if res.Databases == nil {
		return nil, err
	}
	// A type alias makes a slice copy and conversions unnecessary.
	return res.Databases, nil
}

func (cl *Client) Query(ctx context.Context, dbid, query string) ([]map[string]any, error) {
	cmd := &jsonrpc.QueryRequest{
		DBID:  dbid,
		Query: query,
	}
	res := &jsonrpc.QueryResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodQuery), cmd, res)
	if err != nil {
		return nil, err
	}
	return unmarshalMapResults(res.Result)
}

func (cl *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	cmd := &jsonrpc.TxQueryRequest{
		TxHash: txHash,
	}
	res := &jsonrpc.TxQueryResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodTxQuery), cmd, res)
	if err != nil {
		return nil, err
	}

	return &transactions.TcTxQueryResponse{
		Hash:     res.Hash,
		Height:   res.Height,
		Tx:       res.Tx,
		TxResult: *res.TxResult,
	}, nil
}
