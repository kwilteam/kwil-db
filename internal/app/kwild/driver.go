package kwild

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	ec "github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
	"kwil/internal/pkg/graphql/query"
	escrowTypes "kwil/pkg/chain/contracts/escrow/types"
	"kwil/pkg/client"
	"kwil/pkg/databases"
	grpc "kwil/pkg/grpc/client"
	"kwil/pkg/log"
	"math/big"
	"strings"
)

// KwildDriver is a grpc driver for  integration tests
type KwildDriver struct {
	clt         *client.KwilClient
	pk          *ecdsa.PrivateKey
	gatewayAddr string // to ignore the gatewayAddr returned by the config.service

	logger log.Logger
}

func NewKwildDriver(clt *client.KwilClient, pk *ecdsa.PrivateKey, gatewayAddr string, logger log.Logger) *KwildDriver {
	return &KwildDriver{
		clt:         clt,
		pk:          pk,
		gatewayAddr: gatewayAddr,
		logger:      logger,
	}
}

func (d *KwildDriver) GetUserAddress() string {
	return ec.PubkeyToAddress(d.pk.PublicKey).Hex()
}

func (d *KwildDriver) GetServiceConfig(ctx context.Context) (grpc.SvcConfig, error) {
	return d.clt.GetServiceConfig(ctx)
}

func (d *KwildDriver) DepositFund(ctx context.Context, amount *big.Int) error {
	escrow, err := d.clt.EscrowContract(ctx)
	if err != nil {
		return fmt.Errorf("failed to get escrow contract: %w", err)
	}

	_, err = escrow.Deposit(ctx, &escrowTypes.DepositParams{
		Amount:    amount,
		Validator: d.clt.ProviderAddress,
	}, d.pk)
	if err != nil {
		return fmt.Errorf("failed to send deposit transaction: %w", err)
	}

	d.logger.Debug("deposit fund", zap.String("from", d.GetUserAddress()),
		zap.String("to", d.clt.ProviderAddress), zap.String("amount", amount.String()))
	return nil
}

func (d *KwildDriver) GetDepositBalance(ctx context.Context) (*big.Int, error) {
	escrowCtr, err := d.clt.EscrowContract(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get escrow contract: %w", err)
	}

	balanceRes, err := escrowCtr.Balance(ctx, &escrowTypes.DepositBalanceParams{
		Validator: d.clt.ProviderAddress,
		Address:   d.GetUserAddress(),
	})
	if err != nil {
		return nil, err
	}
	return balanceRes.Balance, nil
}

func (d *KwildDriver) ApproveToken(ctx context.Context, spender string, amount *big.Int) error {
	tokenCtr, err := d.clt.TokenContract(ctx)
	if err != nil {
		return err
	}

	_, err = tokenCtr.Approve(ctx, d.clt.EscrowContractAddress, amount, d.pk)
	if err != nil {
		return err
	}
	d.logger.Debug("approve token", zap.String("from", ec.PubkeyToAddress(d.pk.PublicKey).Hex()),
		zap.String("spender", d.clt.EscrowContractAddress), zap.String("amount", amount.String()))
	return nil
}

func (d *KwildDriver) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	tokenCtr, err := d.clt.TokenContract(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting token contract: %w", err)
	}

	allowance, err := tokenCtr.Allowance(d.GetUserAddress(), d.clt.EscrowContractAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting allowance: %w", err)
	}

	return allowance, nil
}

func (d *KwildDriver) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error {
	_, err := d.clt.DeployDatabase(ctx, db, d.pk)
	if err != nil {
		return fmt.Errorf("error deploying database: %w", err)
	}
	d.logger.Debug("deploy database", zap.String("name", db.Name), zap.String("owner", db.Owner))
	return nil
}

func (d *KwildDriver) DatabaseShouldExists(ctx context.Context, owner string, dbName string) error {
	schema, err := d.clt.GetSchema(ctx, owner, dbName)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}

	if strings.EqualFold(schema.Owner, owner) && strings.EqualFold(schema.Name, dbName) {
		return nil
	}
	return fmt.Errorf("database does not exist")
}

func (d *KwildDriver) ExecuteQuery(ctx context.Context, dbName string, queryName string, queryInputs []string) error {
	dbId := databases.GenerateSchemaId(d.GetUserAddress(), dbName)
	qry, err := d.clt.GetQuerySignature(ctx, dbId, queryName)
	if err != nil {
		return fmt.Errorf("error getting query signature: %w", err)
	}

	stringInputs := make(map[string]string) // maps the arg name to the arg value
	for i := 0; i < len(queryInputs); i = i + 2 {
		stringInputs[strings.ToLower(queryInputs[i])] = queryInputs[i+1]
	}
	inputs, err := qry.ConvertInputs(stringInputs)
	if err != nil {
		return fmt.Errorf("error converting inputs: %w", err)
	}

	_, err = d.clt.ExecuteDatabaseById(ctx, dbId, queryName, inputs, d.pk)
	if err != nil {
		return fmt.Errorf("error executing database: %w", err)
	}

	d.logger.Debug("execute query", zap.String("database", dbName), zap.String("query", queryName))
	return nil
}

func (d *KwildDriver) DropDatabase(ctx context.Context, dbName string) error {
	_, err := d.clt.DropDatabase(ctx, dbName, d.pk)
	if err != nil {
		return fmt.Errorf("error dropping database: %w", err)
	}
	d.logger.Debug("drop database", zap.String("name", dbName), zap.String("owner", d.GetUserAddress()))
	return nil
}

func (d *KwildDriver) QueryDatabase(ctx context.Context, queryStr string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/graphql", d.gatewayAddr)
	return query.Query(ctx, url, queryStr)
}
