// package http implements an http transport for the Kwil txsvc client.
package http2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"context"
	"math/big"

	"github.com/antihax/optional"
	"github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	httpTx "github.com/kwilteam/kwil-db/core/rpc/http/tx"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type Client struct {
	conn *httpTx.APIClient
	url  *url.URL
}

// NewClient creates a new http client for the Kwil user service
func NewClient(target *url.URL, opts ...ClientOption) *Client {
	c := &Client{
		url: target,
	}

	clientOpts := &clientOptions{
		client: &http.Client{},
	}

	for _, o := range opts {
		o(clientOpts)
	}

	cfg := httpTx.NewConfiguration()
	cfg.HTTPClient = clientOpts.client
	cfg.BasePath = target.String()
	cfg.Host = target.Host
	cfg.Scheme = target.Scheme

	c.conn = httpTx.NewAPIClient(cfg)

	return c
}

var _ user.TxSvcClient = (*Client)(nil)

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceBroadcast(ctx, httpTx.TxBroadcastRequest{
		Tx: convertTx(tx),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decodedTxHash, err := base64.StdEncoding.DecodeString(result.TxHash)
	if err != nil {
		return nil, err
	}

	return decodedTxHash, nil
}

func (c *Client) Call(ctx context.Context, msg *transactions.CallMessage, opts ...client.ActionCallOption) ([]map[string]any, error) {
	sender := ""
	if msg.Sender != nil {
		sender = base64.StdEncoding.EncodeToString(msg.Sender)
	}

	result, res, err := c.conn.TxServiceApi.TxServiceCall(ctx, httpTx.TxCallRequest{
		AuthType: msg.AuthType,
		Sender:   base64.StdEncoding.EncodeToString(msg.Sender),
		Body: &httpTx.TxCallRequestBody{
			Payload: sender,
		},
	})
	if err != nil {
		// we need to account for a 401 Unauthorized error in this function,
		// but the codegen will return 400 responses as an err, causing this
		// to return. We need to check for this error here and wrap it in
		// our own error type.
		if res != nil && res.StatusCode == http.StatusUnauthorized {
			err = errors.Join(err, client.ErrUnauthorized)
		}

		return nil, err
	}
	defer res.Body.Close()

	// result is []map[string]any encoded in base64
	decodedResult, err := base64.StdEncoding.DecodeString(result.Result)
	if err != nil {
		return nil, err
	}

	// unmashal result
	var resultSet []map[string]any
	err = json.Unmarshal(decodedResult, &resultSet)
	if err != nil {
		return nil, err
	}

	return resultSet, nil
}

func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceChainInfo(ctx)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	parsedHeight, err := strconv.ParseUint(result.Height, 10, 64)
	if err != nil {
		return nil, err
	}

	return &types.ChainInfo{
		ChainID:     result.ChainId,
		BlockHeight: parsedHeight,
		BlockHash:   result.Hash,
	}, nil
}

func (c *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceEstimatePrice(ctx, httpTx.TxEstimatePriceRequest{
		Tx: convertTx(tx),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse result.Price to big.Int
	price, ok := new(big.Int).SetString(result.Price, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse price to big.Int. received: %s", result.Price)
	}

	return price, nil
}

func (c *Client) GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error) {
	// we need to use b64url since this method uses a query string
	result, res, err := c.conn.TxServiceApi.TxServiceGetAccount(ctx, base64.URLEncoding.EncodeToString(pubKey), &httpTx.TxServiceApiTxServiceGetAccountOpts{
		Status: optional.NewString(strconv.FormatUint(uint64(status), 10)), // does not seem to properly handle optional enum properly. This could cause a bug, will need to investigate
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// parse result.Balance to big.Int
	balance, ok := new(big.Int).SetString(result.Account.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance to big.Int. received: %s", result.Account.Balance)
	}

	parsedNonce, err := strconv.ParseInt(result.Account.Nonce, 10, 64)
	if err != nil {
		return nil, err
	}

	return &types.Account{
		Identifier: pubKey,
		Balance:    balance,
		Nonce:      parsedNonce,
	}, nil
}

func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceGetSchema(ctx, dbid)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	convertedSchema, err := convertHttpSchema(result.Schema)
	if err != nil {
		return nil, err
	}

	return convertedSchema, nil
}

func (c *Client) ListDatabases(ctx context.Context, ownerPubKey []byte) ([]string, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceListDatabases(ctx, base64.StdEncoding.EncodeToString(ownerPubKey))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return result.Databases, nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	result, res, err := c.conn.TxServiceApi.TxServicePing(ctx, &httpTx.TxServiceApiTxServicePingOpts{
		Message: optional.NewString("ping"), // we don't really need this I believe?
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	return result.Message, nil
}

func (c *Client) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceQuery(ctx, httpTx.TxQueryRequest{
		Dbid:  dbid,
		Query: query,
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// result is []map[string]any encoded in base64
	decodedResult, err := base64.StdEncoding.DecodeString(result.Result)
	if err != nil {
		return nil, err
	}

	// unmashal result
	var resultSet []map[string]any
	err = json.Unmarshal(decodedResult, &resultSet)
	if err != nil {
		return nil, err
	}

	return resultSet, nil
}

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	result, res, err := c.conn.TxServiceApi.TxServiceTxQuery(ctx, httpTx.TxTxQueryRequest{
		TxHash: base64.StdEncoding.EncodeToString(txHash),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decodedHeight, err := strconv.ParseInt(result.Height, 10, 64)
	if err != nil {
		return nil, err
	}

	decodedHash, err := base64.StdEncoding.DecodeString(result.Hash)
	if err != nil {
		return nil, err
	}

	convertedTx, err := convertHttpTx(result.Tx)
	if err != nil {
		return nil, err
	}

	convertedTxResult, err := convertHttpTxResult(result.TxResult)
	if err != nil {
		return nil, err
	}

	return &transactions.TcTxQueryResponse{
		Hash:     decodedHash,
		Height:   decodedHeight,
		Tx:       *convertedTx,
		TxResult: *convertedTxResult,
	}, nil
}
