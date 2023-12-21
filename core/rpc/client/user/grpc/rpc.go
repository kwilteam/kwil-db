package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	rpcClient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func convertError(err error) error {
	grpcErr := status.Convert(err)
	if grpcErr == nil {
		return err
	}
	rpcErr := &rpcClient.RPCError{
		Msg:  grpcErr.Message(),
		Code: int32(grpcErr.Code()),
	}
	switch grpcErr.Code() {
	case codes.NotFound:
		return errors.Join(rpcErr, rpcClient.ErrNotFound)
	default:
		return rpcErr
	}
}

func (c *Client) GetAccount(ctx context.Context, identifier []byte, status types.AccountStatus) (*types.Account, error) {
	pbStatus := txpb.AccountStatus(status)
	res, err := c.TxClient.GetAccount(ctx, &txpb.GetAccountRequest{
		Identifier: identifier,
		Status:     &pbStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	bigBal, ok := new(big.Int).SetString(res.Account.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance")
	}

	acc := &types.Account{
		Identifier: res.Account.Identifier,
		Balance:    bigBal,
		Nonce:      res.Account.Nonce,
	}

	return acc, nil
}

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	res, err := c.TxClient.TxQuery(ctx, &txpb.TxQueryRequest{
		TxHash: txHash,
	})
	if err != nil {
		return nil, convertError(err)
	}

	return conversion.ConvertFromPBTxQueryResp(res)
}

// ChainInfo gets information on the blockchain of the remote host.
func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	res, err := c.TxClient.ChainInfo(ctx, &txpb.ChainInfoRequest{})
	if err != nil {
		return nil, err
	}
	return &types.ChainInfo{
		ChainID:     res.ChainId,
		BlockHeight: res.Height,
		BlockHash:   res.Hash,
	}, nil
}

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error) {
	pbTx := convertTx(tx)
	res, err := c.TxClient.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		statusError, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("unrecognized broadcast failure: %w", err)
		}

		code, message := statusError.Code(), statusError.Message()
		err = fmt.Errorf("%v (%d)", message, code)

		for _, detail := range statusError.Details() {
			if bcastErr, ok := detail.(*txpb.BroadcastErrorDetails); ok {
				switch txCode := transactions.TxCode(bcastErr.Code); txCode {
				case transactions.CodeWrongChain:
					err = errors.Join(err, transactions.ErrWrongChain)
				case transactions.CodeInvalidNonce:
					err = errors.Join(err, transactions.ErrInvalidNonce)
				default:
					err = errors.Join(err, errors.New(txCode.String()))
				}
			} else { // else unknown details type
				err = errors.Join(err, fmt.Errorf("unrecognized status error detail type %T", detail))
			}
		}
		return nil, err
	}

	return res.GetTxHash(), nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	res, err := c.TxClient.Ping(ctx, &txpb.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}

func (c *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	// convert transaction to proto
	pbTx := convertTx(tx)

	res, err := c.TxClient.EstimatePrice(ctx, &txpb.EstimatePriceRequest{
		Tx: pbTx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate cost: %w", err)
	}

	bigCost, ok := new(big.Int).SetString(res.Price, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert price to big.Int")
	}

	return bigCost, nil
}

func (c *Client) Call(ctx context.Context, req *transactions.CallMessage,
	_ ...rpcClient.ActionCallOption) ([]map[string]any, error) {
	var sender []byte
	if req.Sender != nil {
		sender = req.Sender
	}

	callReq := &txpb.CallRequest{
		Body: &txpb.CallRequest_Body{
			Payload: req.Body.Payload,
		},
		AuthType: req.AuthType,
		Sender:   sender,
	}

	res, err := c.TxClient.Call(ctx, callReq)

	if err != nil {
		return nil, fmt.Errorf("failed to call: %w", err)
	}

	return unmarshalMapResults(res.Result)
}

func (c *Client) GetConfig(ctx context.Context) (*SvcConfig, error) {
	res, err := c.TxClient.GetConfig(ctx, &txpb.GetConfigRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return &SvcConfig{
		ChainCode:       res.ChainCode,
		PoolAddress:     res.PoolAddress,
		ProviderAddress: res.ProviderAddress,
	}, nil
}

type SvcConfig struct {
	ChainCode       int64
	PoolAddress     string
	ProviderAddress string
}

func (c *Client) ListDatabases(ctx context.Context, ownerIdentifier []byte) ([]*types.DatasetIdentifier, error) {
	res, err := c.TxClient.ListDatabases(ctx, &txpb.ListDatabasesRequest{
		Owner: ownerIdentifier,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	datasets := make([]*types.DatasetIdentifier, len(res.Databases))
	for i, db := range res.Databases {
		datasets[i] = &types.DatasetIdentifier{
			Name:  db.Name,
			Owner: db.Owner,
			DBID:  db.Dbid,
		}
	}

	return datasets, nil
}

func (c *Client) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	res, err := c.TxClient.Query(ctx, &txpb.QueryRequest{
		Dbid:  dbid,
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return unmarshalMapResults(res.Result)
}

func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	res, err := c.TxClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	return conversion.ConvertFromPBSchema(res.Schema), nil
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

	return result, nil
}
