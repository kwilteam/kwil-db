package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	cc "kwil/pkg/chain/client"
	ccs "kwil/pkg/chain/client/service"
	"kwil/pkg/chain/contracts/escrow"
	"kwil/pkg/chain/contracts/token"
	chainCodes "kwil/pkg/chain/types"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
	grpcClient "kwil/pkg/grpc/client/v1"
	kTx "kwil/pkg/tx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client          *grpcClient.Client
	datasets        map[string]*models.Dataset
	PrivateKey      *ecdsa.PrivateKey
	chainCode       chainCodes.ChainCode
	ProviderAddress string
	PoolAddress     string
	usingProvider   bool
	chainRpcUrl     string
	chainClient     cc.ChainClient
	tokenContract   token.TokenContract
	tokenAddress    string
	poolContract    escrow.EscrowContract
}

// New creates a new client
func New(ctx context.Context, target string, opts ...ClientOpt) (c *Client, err error) {
	c = &Client{
		datasets:        make(map[string]*models.Dataset),
		chainCode:       chainCodes.LOCAL,
		ProviderAddress: "",
		PoolAddress:     "",
		usingProvider:   true,
		chainRpcUrl:     "",
	}

	for _, opt := range opts {
		opt(c)
	}

	defer func() {
		if c.chainRpcUrl != "" {
			e := c.initChainClient(ctx)
			if err != nil {
				err = e
			}
		}
	}()

	if !c.usingProvider {
		return c, nil
	}

	c.client, err = grpcClient.New(target, grpc.WithTransportCredentials(
		insecure.NewCredentials(), // TODO: should add client configuration for secure transport
	))
	if err != nil {
		return nil, err
	}

	config, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	c.ProviderAddress = config.ProviderAddress
	c.PoolAddress = config.PoolAddress
	c.chainCode = chainCodes.ChainCode(config.ChainCode)

	// re-apply opts to override provider config
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (c *Client) initChainClient(ctx context.Context) error {
	if c.chainRpcUrl == "" {
		return fmt.Errorf("chain rpc url is not set")
	}

	var err error
	c.chainClient, err = ccs.NewChainClient(c.chainRpcUrl,
		ccs.WithChainCode(c.chainCode),
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
	if c.tokenAddress == "" {
		err := c.initPoolContract(ctx)
		if err != nil {
			return fmt.Errorf("failed to init pool contract to get token address: %w", err)
		}
	}

	var err error
	c.tokenContract, err = c.chainClient.Contracts().Token(c.tokenAddress)
	if err != nil {
		return fmt.Errorf("failed to create token contract: %w", err)
	}

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

	c.tokenAddress = c.poolContract.TokenAddress()

	return nil
}

// GetSchema returns the schema of a database
func (c *Client) GetSchema(ctx context.Context, dbid string) (*models.Dataset, error) {
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
func (c *Client) DeployDatabase(ctx context.Context, ds *models.Dataset) (*kTx.Receipt, error) {
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
func (c *Client) deploySchemaTx(ctx context.Context, ds *models.Dataset) (*kTx.Transaction, error) {
	return c.newTx(ctx, kTx.DEPLOY_DATABASE, ds)
}

// DropDatabase drops a database
func (c *Client) DropDatabase(ctx context.Context, name string) (*kTx.Receipt, error) {
	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	tx, err := c.dropDatabaseTx(ctx, &models.DatasetIdentifier{
		Owner: address,
		Name:  name,
	})
	if err != nil {
		return nil, err
	}

	return c.client.Broadcast(ctx, tx)
}

// dropDatabaseTx creates a new transaction to drop a database
func (c *Client) dropDatabaseTx(ctx context.Context, dbIdent *models.DatasetIdentifier) (*kTx.Transaction, error) {
	return c.newTx(ctx, kTx.DROP_DATABASE, dbIdent)
}

// ExecuteAction executes an action
func (c *Client) ExecuteAction(ctx context.Context, dbid string, action string, inputs []map[string]any) (*kTx.Receipt, error) {
	encodedValues, err := encodeInputs(inputs)
	if err != nil {
		return nil, err
	}

	executionBody := &models.ActionExecution{
		Action: action,
		DBID:   dbid,
		Params: encodedValues,
	}

	tx, err := c.executeActionTx(ctx, executionBody)
	if err != nil {
		return nil, err
	}

	return c.client.Broadcast(ctx, tx)
}

// executeActionTx creates a new transaction to execute an action
func (c *Client) executeActionTx(ctx context.Context, executionBody *models.ActionExecution) (*kTx.Transaction, error) {
	return c.newTx(ctx, kTx.EXECUTE_ACTION, executionBody)
}

// encodeInputs converts an input map to a map of encoded values
func encodeInputs(inputs []map[string]any) ([]map[string][]byte, error) {
	encoded := make([]map[string][]byte, 0)
	for _, record := range inputs {
		encodedRecord := make(map[string][]byte)
		for k, v := range record {
			encodedValue, err := types.New(v)
			if err != nil {
				return nil, err
			}
			encodedRecord[k] = encodedValue.Bytes()
		}

		encoded = append(encoded, encodedRecord)
	}
	return encoded, nil
}

// GetConfig returns the provider config
func (c *Client) GetConfig(ctx context.Context) (*grpcClient.SvcConfig, error) {
	return c.client.GetConfig(ctx)
}
