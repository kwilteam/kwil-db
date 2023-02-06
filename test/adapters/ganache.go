package adapters

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"kwil/pkg/chain/client/dto"
	"kwil/pkg/chain/types"
	"kwil/pkg/fund"
	ethFund "kwil/pkg/fund/ethereum"
	"kwil/pkg/logger"
	"kwil/test/utils/deployer"
	"kwil/test/utils/deployer/eth-deployer"
	"math/big"
	"testing"
	"time"
)

const (
	GanachePort = "8545"

	WalletMnemonic    = "test test test test test test test test test test test junk"
	WalletHDPath      = "m/44'/60'/0'"
	DeployerAccount   = "0x1e59ce931B4CFea3fe4B875411e280e173cB7A9C"
	UserAccount       = "0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7"
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
		Image:        "trufflesuite/ganache:v7.7.3",
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

func getChainEndpoint(t *testing.T, ctx context.Context, _chainCode types.ChainCode) (exposedEndpoint string, unexposedEndpoint string) {
	// create ganache(pretend to be Goerli testnet) container
	var err error
	ganacheDocker := StartGanacheDockerService(t, ctx, _chainCode.ToChainId().String())
	exposedEndpoint, err = ganacheDocker.ExposedEndpoint(ctx)
	require.NoError(t, err, "failed to get exposed endpoint")
	unexposedEndpoint, err = ganacheDocker.UnexposedEndpoint(ctx)
	require.NoError(t, err, "failed to get unexposed endpoint")
	return exposedEndpoint, unexposedEndpoint
}

func GetChainDriverAndDeployer(t *testing.T, ctx context.Context, rpcUrl string, deployerPrivateKey string,
	_chainCode types.ChainCode, userPrivateKey string, fundingPoolAddress string, domination *big.Int, logger logger.Logger) (*ethFund.Driver, deployer.Deployer, *fund.Config, map[string]string) {
	userPK, err := crypto.HexToECDSA(userPrivateKey)
	require.NoError(t, err)

	if rpcUrl != "" {
		userFundConfig := fund.Config{
			Chain: dto.Config{
				ChainCode:         int64(_chainCode),
				RpcUrl:            rpcUrl,
				BlockConfirmation: 10,
				ReconnectInterval: 30,
			},
			Wallet:      userPK,
			PoolAddress: fundingPoolAddress,
		}
		t.Logf("create chain driver to %s", rpcUrl)
		chainDriver := ethFund.New(rpcUrl, logger)
		chainDriver.SetFundConfig(&userFundConfig)

		t.Logf("create chain deployer to %s", rpcUrl)
		chainDeployer := eth_deployer.NewEthDeployer(rpcUrl, deployerPrivateKey, domination)
		chainDeployer.UpdateContract(ctx, fundingPoolAddress)
		chainEnvs := map[string]string{}

		return chainDriver, chainDeployer, &userFundConfig, chainEnvs
	}

	exposedRpc, unexposedRpc := getChainEndpoint(t, ctx, _chainCode)

	t.Logf("create chain driver to %s", exposedRpc)
	chainDriver := ethFund.New(exposedRpc, logger)
	t.Logf("create chain deployer to %s", exposedRpc)
	chainDeployer := eth_deployer.NewEthDeployer(exposedRpc, deployerPrivateKey, domination)
	tokenAddress, err := chainDeployer.DeployToken(ctx)
	require.NoError(t, err, "failed to deploy token")
	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress.String())
	require.NoError(t, err, "failed to deploy escrow")

	// to be used by kwild container
	chainEnvs := map[string]string{
		"KWIL_FUND_RPC_URL":            unexposedRpc, //kwild will call using docker network
		"KWIL_FUND_POOL_ADDRESS":       escrowAddress.String(),
		"KWIL_FUND_WALLET":             deployerPrivateKey,
		"KWIL_FUND_CHAIN_CODE":         fmt.Sprintf("%d", _chainCode),
		"KWIL_FUND_BLOCK_CONFIRMATION": "1",
		"KWIL_FUND_RECONNECT_INTERVAL": "30",
	}

	userFundConfig := &fund.Config{
		Wallet:           userPK,
		TokenAddress:     tokenAddress.String(),
		PoolAddress:      escrowAddress.String(),
		ValidatorAddress: chainDeployer.Account.String(),
		Chain: dto.Config{
			ChainCode:         int64(_chainCode),
			RpcUrl:            exposedRpc,
			BlockConfirmation: 10,
			ReconnectInterval: 30,
		},
	}

	chainDriver.SetFundConfig(userFundConfig)

	return chainDriver, chainDeployer, userFundConfig, chainEnvs
}
