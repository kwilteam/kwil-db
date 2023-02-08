package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"kwil/pkg/chain/client/dto"
	"kwil/pkg/chain/types"
	"kwil/pkg/fund"
	ethFund "kwil/pkg/fund/ethereum"
	"kwil/pkg/log"
	"kwil/test/utils/deployer"
	"kwil/test/utils/deployer/eth-deployer"
	"math/big"
	"testing"
	"time"
)

const (
	GanachePort  = "8545"
	GanacheImage = "trufflesuite/ganache:v7.7.3"

	WalletMnemonic    = "test test test test test test test test test test test junk"
	WalletHDPath      = "m/44'/60'/0'"
	DeployerAccountPK = "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5"
	UserAccountPK     = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
)

// ganacheContainer represents the ganache container type used in the module
type ganacheContainer struct {
	TContainer
}

func (c *ganacheContainer) ExposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.ExposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "ws://" + endpoint, nil
}

func (c *ganacheContainer) UnexposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.UnexposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "ws://" + endpoint, nil
}

// setupGanache creates an instance of the ganache container type
func setupGanache(ctx context.Context, chainId string, opts ...containerOption) (*ganacheContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("ganache-%d", time.Now().Unix()),
		Image:        GanacheImage,
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		Networks:     []string{"test-network"},
		ExposedPorts: []string{},
		Cmd: []string{`--wallet.hdPath`, WalletHDPath,
			`--wallet.mnemonic`, WalletMnemonic,
			`--chain.chainId`, chainId},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &ganacheContainer{TContainer{
		Container:     container,
		ContainerPort: GanachePort,
	}}, nil
}

func StartGanacheDockerService(t *testing.T, ctx context.Context, chainId string) *ganacheContainer {
	//t.Helper()

	container, err := setupGanache(ctx,
		chainId,
		WithNetwork(kwilTestNetworkName),
		WithExposedPort(GanachePort),
		WithWaitStrategy(
			wait.ForLog("RPC Listening on 0.0.0.0:8545")))

	require.NoError(t, err, "Could not start ganache container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop ganache container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)

	return container
}

func getChainEndpoint(ctx context.Context, t *testing.T, _chainCode types.ChainCode) (exposedEndpoint string, unexposedEndpoint string) {
	// create ganache(pretend to be Goerli testnet) container
	var err error
	ganacheDocker := StartGanacheDockerService(t, ctx, _chainCode.ToChainId().String())
	exposedEndpoint, err = ganacheDocker.ExposedEndpoint(ctx)
	require.NoError(t, err, "failed to get exposed endpoint")
	unexposedEndpoint, err = ganacheDocker.UnexposedEndpoint(ctx)
	require.NoError(t, err, "failed to get unexposed endpoint")
	return exposedEndpoint, unexposedEndpoint
}

func GetChainDriverAndDeployer(ctx context.Context, t *testing.T, remoteRPCURL string, deployerPKStr string,
	_chainCode types.ChainCode, domination *big.Int, fundCfg *fund.Config, logger log.Logger) (
	*ethFund.Driver, deployer.Deployer, *fund.Config, map[string]string) {
	if remoteRPCURL != "" {
		t.Logf("create chain driver to %s", remoteRPCURL)
		chainDriver := ethFund.New(remoteRPCURL, logger)
		chainDriver.SetFundConfig(fundCfg)

		t.Logf("create chain deployer to %s", remoteRPCURL)
		chainDeployer := eth_deployer.NewEthDeployer(remoteRPCURL, deployerPKStr, domination)
		if err := chainDeployer.UpdateContract(ctx, fundCfg.PoolAddress); err != nil {
			t.Fatalf("failed to update contract: %v", err)
		}
		fundEnvs := map[string]string{}

		return chainDriver, chainDeployer, fundCfg, fundEnvs
	}

	exposedRPC, unexposedRPC := getChainEndpoint(ctx, t, _chainCode)

	t.Logf("create chain driver to %s", exposedRPC)
	chainDriver := ethFund.New(exposedRPC, logger)
	t.Logf("create chain deployer to %s", exposedRPC)
	chainDeployer := eth_deployer.NewEthDeployer(exposedRPC, deployerPKStr, domination)
	tokenAddress, err := chainDeployer.DeployToken(ctx)
	require.NoError(t, err, "failed to deploy token")
	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress.String())
	require.NoError(t, err, "failed to deploy escrow")

	// to be used by kwil container
	fundEnvs := map[string]string{
		"KWILD_FUND_RPC_URL":            unexposedRPC, // kwil will call using docker network
		"KWILD_FUND_POOL_ADDRESS":       escrowAddress.String(),
		"KWILD_FUND_WALLET":             deployerPKStr,
		"KWILD_FUND_CHAIN_CODE":         fmt.Sprintf("%d", _chainCode),
		"KWILD_FUND_BLOCK_CONFIRMATION": "1",
		"KWILD_FUND_RECONNECT_INTERVAL": "30",
	}

	userFundConfig := &fund.Config{
		Wallet:           fundCfg.Wallet,
		TokenAddress:     tokenAddress.String(),
		PoolAddress:      escrowAddress.String(),
		ValidatorAddress: chainDeployer.Account.String(),
		Chain: dto.Config{
			ChainCode:         int64(_chainCode),
			RpcUrl:            exposedRPC,
			BlockConfirmation: 10,
			ReconnectInterval: 30,
		},
	}

	chainDriver.SetFundConfig(userFundConfig)

	return chainDriver, chainDeployer, userFundConfig, fundEnvs
}
