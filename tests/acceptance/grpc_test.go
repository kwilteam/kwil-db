package acceptance_test

import (
	"context"
	"fmt"
	"kwil/tests/adapters"
	"kwil/tests/specifications"
	"kwil/tests/utils/deployer"
	"kwil/x/chain/types"
	"kwil/x/types/databases"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	DeployerPrivateKeyName = "deployer-pk"
	UserPrivateKeyName     = "user-pk"
)

func keepMiningBlocks(ctx context.Context, deployer deployer.Deployer, account string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(5 * time.Second)
			// to mine new blocks
			err := deployer.FundAccount(ctx, account, 1)
			if err != nil {
				fmt.Println("funded user account failed", err)
			}
		}
	}
}

func TestGrpcServerDatabaseService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// test user
	userAddr := adapters.UserAccount
	viper.Set(types.PrivateKeyFlag, adapters.UserAccountPK)
	// test deployer
	viper.Set(DeployerPrivateKeyName, adapters.DeployerAccountPK)
	// database schema
	dbSchemaPath := "./test-data/database_schema.json"

	// config
	// set blow test against a real grpc server
	// and a different private key for the user,
	// and different database_schema.json
	remoteKwildAddr := ""
	providerEndpoint := ""

	t.Run("should approve token", func(t *testing.T) {
		ctx := context.Background()
		// setup
		chainDriver, deployer, _ := adapters.GetChainDriverAndDeployer(t, ctx, providerEndpoint, viper.GetString(DeployerPrivateKeyName))

		// Given user is funded
		err := deployer.FundAccount(ctx, userAddr, 200)
		assert.NoError(t, err, "failed to fund user account")

		// and user has approved funding_pool to spend his funds
		specifications.ApproveTokenSpecification(t, ctx, chainDriver)
	})

	t.Run("should deposit fund", func(t *testing.T) {
		ctx := context.Background()
		// setup
		chainDriver, deployer, _ := adapters.GetChainDriverAndDeployer(
			t, ctx, providerEndpoint, viper.GetString(DeployerPrivateKeyName))

		// Given user is funded with escrow token
		err := deployer.FundAccount(ctx, userAddr, 200)
		assert.NoError(t, err, "failed to fund user account")

		// and user has approved funding_pool to spend his token
		specifications.ApproveTokenSpecification(t, ctx, chainDriver)

		// should be able to deposit fund
		specifications.DepositFundSpecification(t, ctx, chainDriver)
	})

	t.Run("should drop and drop database", func(t *testing.T) {
		ctx := context.Background()
		// setup
		specifications.SetSchemaLoader(
			&specifications.FileDatabaseSchemaLoader{
				FilePath: dbSchemaPath,
				Modifier: func(db *databases.Database[[]byte]) {
					db.Owner = userAddr
				}})
		chainDriver, deployer, kwildEnvs := adapters.GetChainDriverAndDeployer(
			t, ctx, providerEndpoint, viper.GetString(DeployerPrivateKeyName))

		// Given user is funded
		err := deployer.FundAccount(ctx, userAddr, 200)
		assert.NoError(t, err, "failed to fund user account")
		go keepMiningBlocks(ctx, deployer, userAddr)

		// and user has approved funding_pool to spend his funds
		specifications.ApproveTokenSpecification(t, ctx, chainDriver)

		// and user is registered to a validator
		specifications.DepositFundSpecification(t, ctx, chainDriver)

		// When user deployed database
		dbFiles := map[string]string{
			"../../scripts/pg-init-scripts/initdb.sh": "/docker-entrypoint-initdb.d/initdb.sh"}
		driver := adapters.GetGrpcDriver(t, ctx, remoteKwildAddr, dbFiles, kwildEnvs)

		time.Sleep(3 * time.Second) // chain sync
		specifications.DatabaseDeploySpecification(t, ctx, driver)

		// Then user should be able to drop database
		specifications.DatabaseDropSpecification(t, ctx, driver)
	})

	t.Run("should execute database", func(t *testing.T) {
		ctx := context.Background()
		// setup
		specifications.SetSchemaLoader(
			&specifications.FileDatabaseSchemaLoader{
				FilePath: dbSchemaPath,
				Modifier: func(db *databases.Database[[]byte]) {
					db.Owner = userAddr
				}})
		chainDriver, deployer, kwild_envs := adapters.GetChainDriverAndDeployer(
			t, ctx, providerEndpoint, viper.GetString(DeployerPrivateKeyName))

		// Given user is funded
		err := deployer.FundAccount(ctx, userAddr, 200)
		assert.NoError(t, err, "failed to fund user account")
		go keepMiningBlocks(ctx, deployer, userAddr)

		// and user has approved funding_pool to spend his funds
		specifications.ApproveTokenSpecification(t, ctx, chainDriver)

		// and user is registered to a validator
		specifications.DepositFundSpecification(t, ctx, chainDriver)

		// When user deployed database
		dbFiles := map[string]string{
			"../../scripts/pg-init-scripts/initdb.sh": "/docker-entrypoint-initdb.d/initdb.sh"}
		driver := adapters.GetGrpcDriver(t, ctx, remoteKwildAddr, dbFiles, kwild_envs)

		time.Sleep(3 * time.Second) // chain sync
		specifications.DatabaseDeploySpecification(t, ctx, driver)

		// Then user should be able to execute database
		// TODO: separate cases?
		specifications.ExecuteDBInsertSpecification(t, ctx, driver)
		specifications.ExecuteDBUpdateSpecification(t, ctx, driver)
		specifications.ExecuteDBDeleteSpecification(t, ctx, driver)

		// and user should be able to drop database
		specifications.DatabaseDropSpecification(t, ctx, driver)
	})
}
