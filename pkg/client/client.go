package client

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	kwilCrypto "github.com/kwilteam/kwil-db/pkg/crypto"
	"strings"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	cc "github.com/kwilteam/kwil-db/pkg/chain/client"
	ccs "github.com/kwilteam/kwil-db/pkg/chain/client/service"
	"github.com/kwilteam/kwil-db/pkg/chain/contracts/escrow"
	"github.com/kwilteam/kwil-db/pkg/chain/contracts/token"
	chainCodes "github.com/kwilteam/kwil-db/pkg/chain/types"
	grpcClient "github.com/kwilteam/kwil-db/pkg/grpc/client/v1"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client           *grpcClient.Client
	datasets         map[string]*entity.Schema
	PrivateKey       *ecdsa.PrivateKey
	ChainCode        chainCodes.ChainCode
	ProviderAddress  string
	PoolAddress      string
	usingProvider    bool
	withServerConfig bool
	chainRpcUrl      string
	chainClient      cc.ChainClient
	tokenContract    token.TokenContract
	TokenAddress     string
	TokenSymbol      string
	poolContract     escrow.EscrowContract
}

// New creates a new client
func New(ctx context.Context, target string, opts ...ClientOpt) (c *Client, err error) {
	c = &Client{
		datasets:         make(map[string]*entity.Schema),
		ChainCode:        chainCodes.LOCAL,
		ProviderAddress:  "",
		PoolAddress:      "",
		usingProvider:    true,
		withServerConfig: true,
		chainRpcUrl:      "",
		TokenAddress:     "",
		TokenSymbol:      "",
	}

	for _, opt := range opts {
		opt(c)
	}

	defer func(c *Client) {
		if c.chainRpcUrl != "" {
			tempErr := c.initChainClient(ctx)
			if tempErr != nil {
				err = tempErr
			}
		}
	}(c)

	if !c.usingProvider {
		if c.chainRpcUrl != "" {
			e := c.initChainClient(ctx)
			if err != nil {
				err = e
			}
		}
		return c, nil
	}

	c.client, err = grpcClient.New(target, grpc.WithTransportCredentials(
		insecure.NewCredentials(), // TODO: should add client configuration for secure transport
	))
	if err != nil {
		return nil, err
	}

	if c.withServerConfig {
		err = c.loadServerConfig(ctx)
		if err != nil {
			return nil, err
		}
	}

	// re-apply opts to override provider config
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (c *Client) loadServerConfig(ctx context.Context) error {
	config, err := c.GetConfig(ctx)
	if err != nil {
		return err
	}
	c.ProviderAddress = config.ProviderAddress
	c.PoolAddress = config.PoolAddress
	c.ChainCode = chainCodes.ChainCode(config.ChainCode)

	return nil
}

func (c *Client) initChainClient(ctx context.Context) error {
	if c.chainRpcUrl == "" {
		return fmt.Errorf("chain rpc url is not set")
	}

	var err error
	c.chainClient, err = ccs.NewChainClient(c.chainRpcUrl,
		ccs.WithChainCode(c.ChainCode),
	)
	if err != nil {
		return fmt.Errorf("failed to create chain client: %w", err)
	}

	return nil
}

func (c *Client) initTokenContract(ctx context.Context) error {
	if c.chainClient == nil {
		return fmt.Errorf("chain client is not initialized")
	}
	if c.TokenAddress == "" {
		err := c.initPoolContract(ctx)
		if err != nil {
			return fmt.Errorf("failed to init pool contract to get token address: %w", err)
		}
	}

	var err error
	c.tokenContract, err = c.chainClient.Contracts().Token(c.TokenAddress)
	if err != nil {
		return fmt.Errorf("failed to create token contract: %w", err)
	}

	c.TokenSymbol = c.tokenContract.Symbol()

	return nil
}

func (c *Client) initPoolContract(ctx context.Context) error {
	if c.chainClient == nil {
		return fmt.Errorf("chain client is not initialized")
	}
	if c.PoolAddress == "" {
		return fmt.Errorf("pool address is not set")
	}

	var err error
	c.poolContract, err = c.chainClient.Contracts().Escrow(c.PoolAddress)
	if err != nil {
		return fmt.Errorf("failed to create escrow contract: %w", err)
	}

	c.TokenAddress = c.poolContract.TokenAddress()

	return nil
}

// GetSchema returns the entity of a database
func (c *Client) GetSchema(ctx context.Context, dbid string) (*entity.Schema, error) {
	ds, ok := c.datasets[dbid]
	if ok {
		return ds, nil
	}

	ds, err := c.client.GetSchema(ctx, dbid)
	if err != nil {
		return nil, err
	}

	c.datasets[dbid] = ds
	return ds, nil
}

// DeployDatabase deploys a schema
func (c *Client) DeployDatabase(ctx context.Context, ds *entity.Schema) (*kTx.Receipt, error) {
	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	if ds.Owner != address {
		return nil, fmt.Errorf("dataset owner is not the same as the address")
	}

	tx, err := c.deploySchemaTx(ctx, ds)
	if err != nil {
		return nil, err
	}

	return c.client.Broadcast(ctx, tx)
}

// deploySchemaTx creates a new transaction to deploy a schema
func (c *Client) deploySchemaTx(ctx context.Context, ds *entity.Schema) (*kTx.Transaction, error) {
	return c.newTx(ctx, kTx.DEPLOY_DATABASE, ds)
}

// DropDatabase drops a database
func (c *Client) DropDatabase(ctx context.Context, name string) (*kTx.Receipt, error) {
	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	identifier := &datasetIdentifier{
		Owner: address,
		Name:  name,
	}

	tx, err := c.dropDatabaseTx(ctx, identifier)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}

	delete(c.datasets, identifier.Dbid())

	return res, nil
}

// dropDatabaseTx creates a new transaction to drop a database
func (c *Client) dropDatabaseTx(ctx context.Context, dbIdent *datasetIdentifier) (*kTx.Transaction, error) {
	return c.newTx(ctx, kTx.DROP_DATABASE, dbIdent)
}

// ExecuteAction executes an action.
// It returns the receipt, as well as outputs which is the decoded body of the receipt.
func (c *Client) ExecuteAction(ctx context.Context, dbid string, action string, inputs []map[string]any) (*kTx.Receipt, []map[string]any, error) {
	executionBody := &actionExecution{
		Action: action,
		DBID:   dbid,
		Params: inputs,
	}

	tx, err := c.executeActionTx(ctx, executionBody)
	if err != nil {
		return nil, nil, err
	}

	res, err := c.client.Broadcast(ctx, tx)
	if err != nil {
		return nil, nil, err
	}

	outputs, err := decodeOutputs(res.Body)
	if err != nil {
		return nil, nil, err
	}

	return res, outputs, nil
}

func createActionPayload(dbid string, action string, inputs []map[string]any) *txpb.ActionPayload {
	payload := &txpb.ActionPayload{
		Dbid:   dbid,
		Action: action,
		Params: make([]*txpb.ActionInput, len(inputs)),
	}

	for i, param := range inputs {
		input := make(map[string]string)
		for k, v := range param {
			input[k] = fmt.Sprintf("%v", v)
		}
		payload.Params[i] = &txpb.ActionInput{Input: input}
	}

	return payload
}

// CallAction call an action, if auxiliary `mustsign` is set, need to sign the action payload. It returns the records.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs []map[string]any) (*Records, error) {
	// @yaiba
	// NOTE: not just this RPC, similar RPCs should enforce client to get schema json first locally
	// Or, cache the schema json in client side local file system, better UX.
	remoteSchema, err := c.GetSchema(ctx, dbid)
	if err != nil {
		return nil, err
	}

	for _, act := range remoteSchema.Actions {
		if act.Name == action && act.Mutability == entity.MutabilityView.String() {
			// NOTE: here the actionBody to sign is not the same as the actionPayload in the request
			actionPayloadToSign := &actionExecution{
				Action: action,
				DBID:   dbid,
				Params: inputs,
			}

			req := &txpb.CallRequest{
				Payload: createActionPayload(dbid, action, inputs),
				Sender:  kwilCrypto.AddressFromPrivateKey(c.PrivateKey),
			}

			// handle auxiliary setting
			for _, aux := range act.Auxiliaries {
				if aux == entity.AuxiliaryTypeMustSign.String() {
					payloadBytes, err := json.Marshal(actionPayloadToSign)
					if err != nil {
						return nil, fmt.Errorf("failed to serialize data: %w", err)
					}

					payloadHash := kwilCrypto.Sha384(payloadBytes)
					sig, err := kwilCrypto.Sign(payloadHash, c.PrivateKey)
					if err != nil {
						return nil, fmt.Errorf("failed to sign transaction: %v", err)
					}

					req.Signature = &txpb.Signature{
						SignatureBytes: sig.Signature,
						SignatureType:  sig.Type.Int32(),
					}
				}
			}

			res, err := c.client.Call(ctx, req)
			if err != nil {
				return nil, err
			}

			return NewRecordsFromMaps(res), nil
		}
	}

	return nil, fmt.Errorf("action %s is not found", action)
}

func decodeOutputs(bts []byte) ([]map[string]any, error) {
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

// executeActionTx creates a new transaction to execute an action
func (c *Client) executeActionTx(ctx context.Context, executionBody *actionExecution) (*kTx.Transaction, error) {
	return c.newTx(ctx, kTx.EXECUTE_ACTION, executionBody)
}

// GetConfig returns the provider config
func (c *Client) GetConfig(ctx context.Context) (*grpcClient.SvcConfig, error) {
	return c.client.GetConfig(ctx)
}

// Query executes a query
func (c *Client) Query(ctx context.Context, dbid string, query string) (*Records, error) {
	res, err := c.client.Query(ctx, dbid, query)
	if err != nil {
		return nil, err
	}

	return NewRecordsFromMaps(res), nil
}

func (c *Client) ListDatabases(ctx context.Context, owner string) ([]string, error) {
	owner = strings.ToLower(owner)
	return c.client.ListDatabases(ctx, owner)
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.client.Ping(ctx)
}

func (c *Client) GetAccount(ctx context.Context, address string) (*balances.Account, error) {
	return c.client.GetAccount(ctx, address)
}
