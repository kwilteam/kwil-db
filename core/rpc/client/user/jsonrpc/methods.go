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

	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	userjson "github.com/kwilteam/kwil-db/core/rpc/json/user"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	jsonUtil "github.com/kwilteam/kwil-db/core/utils/json"
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
	cmd := &userjson.PingRequest{
		Message: "ping",
	}
	res := &userjson.PingResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodPing), cmd, res)
	if err != nil {
		return "", err
	}
	return res.Message, nil
}

func (cl *Client) Broadcast(ctx context.Context, tx *transactions.Transaction, sync rpcclient.BroadcastWait) ([]byte, error) {
	cmd := &userjson.BroadcastRequest{
		Tx:   tx,
		Sync: (*userjson.BroadcastSync)(&sync),
	}
	res := &userjson.BroadcastResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodBroadcast), cmd, res)
	if err != nil {
		var jsonRPCErr *jsonrpc.Error
		if errors.As(err, &jsonRPCErr) && jsonRPCErr.Code == jsonrpc.ErrorTxExecFailure && len(jsonRPCErr.Data) > 0 {
			var berr userjson.BroadcastError
			jsonErr := json.Unmarshal(jsonRPCErr.Data, &berr)
			if jsonErr != nil {
				return nil, errors.Join(jsonErr, err)
			}

			err = errors.Join(berr, err)

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

func (cl *Client) Call(ctx context.Context, msg *transactions.CallMessage, opts ...rpcclient.ActionCallOption) ([]map[string]any, []string, error) {
	cmd := msg // same underlying type presently
	res := &userjson.CallResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodCall), cmd, res)
	if err != nil {
		return nil, nil, err
	}
	records, err := jsonUtil.UnmarshalMapWithoutFloat[[]map[string]any](res.Result)
	if err != nil {
		return nil, nil, err
	}

	return records, res.Logs, nil
}

func (cl *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	cmd := &userjson.ChainInfoRequest{}
	res := &userjson.ChainInfoResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodChainInfo), cmd, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cl *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	cmd := &userjson.EstimatePriceRequest{
		Tx: tx,
	}
	res := &userjson.EstimatePriceResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodPrice), cmd, res)
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
	cmd := &userjson.AccountRequest{
		Identifier: pubKey,
		Status:     &status,
	}
	res := &userjson.AccountResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodAccount), cmd, res)
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
	cmd := &userjson.SchemaRequest{
		DBID: dbid,
	}
	res := &userjson.SchemaResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodSchema), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Schema, nil
}

func (cl *Client) ListDatabases(ctx context.Context, ownerPubKey []byte) ([]*types.DatasetIdentifier, error) {
	cmd := &userjson.ListDatabasesRequest{
		Owner: ownerPubKey,
	}
	res := &userjson.ListDatabasesResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodDatabases), cmd, res)
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
	cmd := &userjson.QueryRequest{
		DBID:  dbid,
		Query: query,
	}
	res := &userjson.QueryResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodQuery), cmd, res)
	if err != nil {
		return nil, err
	}
	return jsonUtil.UnmarshalMapWithoutFloat[[]map[string]any](res.Result)
}

func (cl *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	cmd := &userjson.TxQueryRequest{
		TxHash: txHash,
	}
	res := &userjson.TxQueryResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodTxQuery), cmd, res)
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

// ListMigrations lists all migrations that have been proposed that are still in the pending state.
func (cl *Client) ListMigrations(ctx context.Context) ([]*types.Migration, error) {
	cmd := &userjson.ListMigrationsRequest{}
	res := &userjson.ListMigrationsResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodListMigrations), cmd, res)
	if err != nil {
		return nil, err
	}

	return res.Migrations, nil
}

// LoadChangesets loads changesets from the node's database at the given height.
func (cl *Client) LoadChangeset(ctx context.Context, height int64, index int64) ([]byte, error) {
	cmd := &userjson.ChangesetRequest{
		Height: height,
		Index:  index,
	}
	res := &userjson.ChangesetsResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodLoadChangeset), cmd, res)
	if err != nil {
		return nil, err
	}

	return res.Changesets, nil
}

// ChangesetMetadata gets metadata about the changesets at the given height.
func (cl *Client) ChangesetMetadata(ctx context.Context, height int64) (numChangesets int64, changesetsSizes []int64, err error) {
	cmd := &userjson.ChangesetMetadataRequest{
		Height: height,
	}
	res := &userjson.ChangesetMetadataResponse{}
	err = cl.CallMethod(ctx, string(userjson.MethodLoadChangesetMetadata), cmd, res)
	if err != nil {
		return -1, nil, err
	}

	if res.Height != height {
		return -1, nil, fmt.Errorf("received incorrect block's metadata: got %d, expected %d", res.Height, height)
	}

	return res.Changesets, res.ChunkSizes, nil
}

// GenesisState returns the genesis state of the chain.
func (cl *Client) GenesisState(ctx context.Context) (*types.MigrationMetadata, error) {
	cmd := &userjson.MigrationMetadataRequest{}
	res := &userjson.MigrationMetadataResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodMigrationMetadata), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Metadata, nil
}

// GenesisSnapshotChunk returns a chunk of the genesis snapshot at the given height and chunkIdx.
func (cl *Client) GenesisSnapshotChunk(ctx context.Context, height uint64, chunkIdx uint32) ([]byte, error) {
	cmd := &userjson.MigrationSnapshotChunkRequest{
		ChunkIndex: chunkIdx,
		Height:     height,
	}
	res := &userjson.MigrationSnapshotChunkResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodMigrationGenesisChunk), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Chunk, nil
}

func (cl *Client) MigrationStatus(ctx context.Context) (*types.MigrationState, error) {
	cmd := &userjson.MigrationStatusRequest{}
	res := &userjson.MigrationStatusResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodMigrationStatus), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Status, nil
}

func (cl *Client) Challenge(ctx context.Context) ([]byte, error) {
	cmd := &userjson.ChallengeRequest{}
	res := &userjson.ChallengeResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodChallenge), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Challenge, nil
}

func (cl *Client) Health(ctx context.Context) (*types.Health, error) {
	cmd := &userjson.HealthRequest{}
	res := &userjson.HealthResponse{}
	err := cl.CallMethod(ctx, string(userjson.MethodHealth), cmd, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
