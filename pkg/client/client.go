package client

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"

	cmtCrypto "github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
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
	BcClient         *rpchttp.HTTP
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
	BcRpcUrl         string
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
		BcRpcUrl:         "tcp://localhost:26657",
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

	c.BcClient, err = rpchttp.New(c.BcRpcUrl, "")
	if err != nil {
		return nil, err
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
func (c *Client) ExecuteAction(ctx context.Context, dbid string, action string, inputs []map[string]any) (*kTx.Receipt, error) {
	executionBody := &actionExecution{
		Action: action,
		DBID:   dbid,
		Params: inputs,
	}

	tx, err := c.executeActionTx(ctx, executionBody)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}

	/* 	outputs, err := DecodeOutputs(res.Body)
	   	if err != nil {
	   		return nil, err
	   	}
	*/
	return res, nil
}

// CallAction call an action, if auxiliary `mustsign` is set, need to sign the action payload. It returns the records.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs map[string]any, opts ...CallOpt) ([]map[string]any, error) {
	callOpts := &callOptions{}

	for _, opt := range opts {
		opt(callOpts)
	}

	payload := &kTx.CallActionPayload{
		DBID:   dbid,
		Action: action,
		Params: inputs,
	}

	var signedMsg *kTx.SignedMessage[*kTx.CallActionPayload]
	shouldSign, err := shouldAuthenticate(c.PrivateKey, callOpts.forceAuthenticated)
	if err != nil {
		return nil, err
	}

	if shouldSign {
		signedMsg, err = kTx.CreateSignedMessage(payload, c.PrivateKey)
	} else {
		signedMsg = kTx.CreateEmptySignedMessage(payload)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create signed message: %w", err)
	}

	return c.client.Call(ctx, (*kTx.CallActionMessage)(signedMsg))
}

// shouldAuthenticate decides whether the client should authenticate or not
// if enforced is not nil, it will be used instead of the default value
// otherwise, if the private key is not nil, it will authenticate
func shouldAuthenticate(privateKey *ecdsa.PrivateKey, enforced *bool) (bool, error) {
	if enforced != nil {
		if !*enforced {
			return false, nil
		}

		if privateKey == nil {
			return false, fmt.Errorf("private key is nil, but authentication is enforced")
		}

		return true, nil
	}

	return privateKey != nil, nil
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

func (c *Client) ApproveValidator(ctx context.Context, approver string, joiner string) ([]byte, error) {
	tx, err := c.NewNodeTx(ctx, kTx.VALIDATOR_APPROVE, joiner, approver)
	if err != nil {
		return nil, err
	}

	bts, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	res, err := c.BcClient.BroadcastTxAsync(ctx, bts)
	if err != nil {
		return nil, err
	}
	return res.Hash, nil
}

func (c *Client) ValidatorJoin(ctx context.Context, joiner string, power int64) ([]byte, error) {
	return c.ValidatorUpdate(ctx, joiner, power, kTx.VALIDATOR_JOIN)
}

func (c *Client) ValidatorLeave(ctx context.Context, joiner string, power int64) ([]byte, error) {
	return c.ValidatorUpdate(ctx, joiner, 0, kTx.VALIDATOR_LEAVE)
}

func (c *Client) ValidatorUpdate(ctx context.Context, joinerPrivKey string, power int64, payloadtype kTx.PayloadType) ([]byte, error) {
	var nodeKey cmtCrypto.PrivKey
	key := fmt.Sprintf(`{"type":"tendermint/PrivKeyEd25519","value":"%s"}`, joinerPrivKey)
	err := cmtjson.Unmarshal([]byte(key), &nodeKey)
	if err != nil {
		return nil, err
	}

	fmt.Println("Node PublicKey: ", nodeKey.PubKey())
	bts, _ := json.Marshal(nodeKey.PubKey())
	fmt.Println("Node PublicKey: ", string(bts))

	validator := &validator{
		PubKey: string(bts),
		Power:  power,
	}

	tx, err := c.NewNodeTx(ctx, payloadtype, validator, joinerPrivKey)
	if err != nil {
		return nil, err
	}

	bts, err = json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	res, err := c.BcClient.BroadcastTxAsync(ctx, bts)
	if err != nil {
		return nil, err
	}
	return res.Hash, nil
}
