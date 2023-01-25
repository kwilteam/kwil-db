package adapters

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"kwil/tests/utils/deployer"
	"kwil/tests/utils/deployer/eth-deployer"
	"kwil/x/chain/types"
	"kwil/x/fund"
	ethFund "kwil/x/fund/ethereum"
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
func setupGanache(ctx context.Context, opts ...containerOption) (*ganacheContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("ganache-%d", time.Now().Unix()),
		Image:        "trufflesuite/ganache:v7.7.3",
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		Networks:     []string{"test-network"},
		ExposedPorts: []string{},
		Cmd: []string{`--wallet.hdPath`, WalletHDPath,
			`--wallet.mnemonic`, WalletMnemonic,
			`--chain.chainId`, types.GOERLI.ToChainId().String()},
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

func GetChainDriverAndDeployer(t *testing.T, ctx context.Context, providerEndpoint string, deployerPrivateKey string) (*ethFund.Driver, deployer.Deployer, map[string]string) {
	t.Helper()

	if providerEndpoint != "" {
		deployer := eth_deployer.NewEthDeployer(providerEndpoint, deployerPrivateKey)

		// use default chain config for kwild
		kwildEnvs := map[string]string{}

		return &ethFund.Driver{Addr: providerEndpoint}, deployer, kwildEnvs
	}

	// create ganache container
	ganacheDocker := StartGanacheDockerService(t, ctx)

	// ganache is mimicing testnet
	viper.Set(types.ChainCodeFlag, int64(types.GOERLI))

	exposedProviderEndpoint, err := ganacheDocker.ExposedEndpoint(ctx)
	assert.NoError(t, err, "failed to get endpoint")
	unexposedProviderEndpoint, err := ganacheDocker.UnexposedEndpoint(ctx)
	assert.NoError(t, err, "failed to get endpoint")

	deployer := eth_deployer.NewEthDeployer(exposedProviderEndpoint, deployerPrivateKey)
	viper.Set(fund.ValidatorAddressFlag, deployer.Account.String())

	// deploy smart contracts
	tokenAddress, err := deployer.DeployToken(ctx)
	assert.NoError(t, err, "failed to deploy token")

	escrowAddress, err := deployer.DeployEscrow(ctx, tokenAddress.String())
	assert.NoError(t, err, "failed to deploy escrow")

	// TODO: make kwil client init from configuration, not from flags
	// set viper vars, for creating chain client
	// maybe set this to .env???
	viper.Set(fund.TokenAddressFlag, tokenAddress.String())
	viper.Set(fund.FundingPoolFlag, escrowAddress.String())
	viper.Set(types.EthProviderFlag, exposedProviderEndpoint)
	viper.Set(types.ReconnectionIntervalFlag, 30)
	viper.Set(types.RequiredConfirmationsFlag, 1)

	// to be used by kwild container
	kwildEnvs := map[string]string{
		"DEPOSITS_PROVIDER_ENDPOINT": unexposedProviderEndpoint, //kwild will call using docker network
		"DEPOSITS_CONTRACT_ADDRESS":  escrowAddress.String(),
		"DEPOSITS_WALLET_KEY":        deployerPrivateKey,
		"DEPOSITS_CHAIN":             "Goerli",
		"CHAIN_CODE":                 fmt.Sprintf("%d", types.GOERLI),
		"DEPOSITS_ENABLED":           "false",
		"REQUIRED_CONFIRMATIONS":     "1",
	}

	t.Logf("create chain driver to %s", exposedProviderEndpoint)
	return &ethFund.Driver{Addr: exposedProviderEndpoint}, deployer, kwildEnvs
}
