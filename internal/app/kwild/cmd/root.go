package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/controller/grpc/healthsvc/v0"
	"github.com/kwilteam/kwil-db/internal/controller/grpc/txsvc/v1"
	"google.golang.org/grpc/health/grpc_health_v1"

	abci "github.com/cometbft/cometbft/abci/types"
	ccfg "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"

	// shorthand for chain client service
	"github.com/kwilteam/kwil-db/pkg/balances"
	chainsyncer "github.com/kwilteam/kwil-db/pkg/balances/chain-syncer"
	"github.com/kwilteam/kwil-db/pkg/log"

	kwildbapp "github.com/kwilteam/kwil-db/internal/abci-apps"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
	"github.com/kwilteam/kwil-db/internal/pkg/gateway/middleware/cors"
	"github.com/kwilteam/kwil-db/internal/pkg/healthcheck"
	simple_checker "github.com/kwilteam/kwil-db/internal/pkg/healthcheck/simple-checker"
	grpcServer "github.com/kwilteam/kwil-db/pkg/grpc/server"

	chainClient "github.com/kwilteam/kwil-db/pkg/chain/client"
	ccService "github.com/kwilteam/kwil-db/pkg/chain/client/service" // shorthand for chain client service
	chainTypes "github.com/kwilteam/kwil-db/pkg/chain/types"
	kwilCrypto "github.com/kwilteam/kwil-db/pkg/crypto"
)

func NewStartCmd() *cobra.Command {
	return startCmd
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "kwil grpc server",
	Long:  "Starts node with Kwild and CometBFT services",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := config.LoadKwildConfig()
		if err != nil {
			return err
		}

		fmt.Println("Chain client config: ", cfg.Deposits.ClientChainRPCURL)
		logger := log.New(cfg.Log)
		logger = *logger.Named("kwild")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		fmt.Printf("Initializing kwil server")
		srv, txSvc, err := initialize_kwil_server(ctx, cfg, logger)
		if err != nil {
			return nil
		}
		fmt.Println("Chain client config: ", cfg.Deposits.ClientChainRPCURL)
		fmt.Printf("Starting kwil server")
		app, err := kwildbapp.NewKwilDbApplication(srv, txSvc.GetExecutor())
		if err != nil {
			return nil
		}

		go func(ctx context.Context) {
			srv.Start(ctx)
		}(ctx)

		fmt.Printf("Starting Tendermint node\n")
		// Start the Tendermint node
		cometNode, err := newCometNode(app)
		if err != nil {
			return nil
		}
		txSvc.BcNode = cometNode

		go func() {
			cometNode.Start()
			defer func() {
				cometNode.Stop()
				cometNode.Wait()
			}()
			fmt.Printf("Waiting for any signals\n")
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			<-c
			fmt.Printf("Stopping CometBFT node\n")
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		return nil
	}}

func init() {
	/*
		defaultConfigPath := filepath.Join("$HOME", config.DefaultConfigDir,
			fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigType))
		RootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file to use (default: '%s')", defaultConfigPath))
	*/

	config.BindFlagsAndEnv(startCmd.PersistentFlags())
}

func initialize_kwil_server(ctx context.Context, cfg *config.KwildConfig, logger log.Logger) (*server.Server, *txsvc.Service, error) {
	fmt.Printf("Building chain client\n: %v", cfg)
	fmt.Println("Chain client config: ", cfg.Deposits.ClientChainRPCURL)
	chainclient, err := buildChainClient(cfg, logger)
	if err != nil {
		fmt.Printf("Failed to build chain client: %v", err)
		return nil, nil, fmt.Errorf("failed to build chain client: %w", err)
	}

	// TODO: Move to CometBFT later? or are these different accounts?
	fmt.Printf("Building account repository\n")
	accountStore, err := buildAccountRepository(logger, cfg)
	if err != nil {
		fmt.Printf("Failed to build account repository: %v", err)
		return nil, nil, fmt.Errorf("failed to build account repository: %w", err)
	}

	fmt.Printf("Building chain syncer\n")
	chainSyncer, err := buildChainSyncer(cfg, chainclient, accountStore, logger)
	if err != nil {
		fmt.Printf("Failed to build chain syncer: %v", err)
		return nil, nil, fmt.Errorf("failed to build chain syncer: %w", err)
	}

	fmt.Printf("Building tx service\n")
	txSvc, err := buildTxSvc(ctx, cfg, accountStore, logger)
	if err != nil {
		fmt.Printf("Failed to build tx service: %v", err)
		return nil, nil, fmt.Errorf("failed to build tx service: %w", err)
	}

	fmt.Printf("Building health service\n")
	healthSvc := buildHealthSvc(logger)

	// Commenting this out as we would be using the CometBFT's endpoint
	//fmt.Printf("Building gateway server\n")
	gw := server.NewGWServer(runtime.NewServeMux(), cfg, logger)
	if err := gw.SetupGrpcSvc(ctx); err != nil {
		fmt.Printf("Failed to setup grpc service: %v", err)
		return nil, nil, err
	}
	fmt.Printf("Setting up http service\n")
	if err := gw.SetupHTTPSvc(ctx); err != nil {
		fmt.Printf("Failed to setup http service: %v", err)
		return nil, nil, err
	}

	fmt.Printf("Adding middlewares\n")
	gw.AddMiddlewares(
		// from innermost middleware
		//auth.MAuth(keyManager, logger),
		cors.MCors([]string{}),
	)

	//grpc server
	grpcServer := grpcServer.New(logger)
	txpb.RegisterTxServiceServer(grpcServer, txSvc)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSvc)
	fmt.Printf("Registering grpc services\n")
	server := &server.Server{
		Cfg:         cfg,
		Log:         logger,
		ChainSyncer: chainSyncer,
		Http:        gw,
		Grpc:        grpcServer,
	}
	return server, txSvc, nil
}

func newCometNode(app abci.Application) (*nm.Node, error) {
	config := ccfg.DefaultConfig()
	CometHomeDir := os.Getenv("COMET_BFT_HOME")
	fmt.Printf("Home Directory: %v", CometHomeDir)
	config.SetRoot(CometHomeDir)

	viper.SetConfigFile(fmt.Sprintf("%s/%s", CometHomeDir, "config/config.toml"))
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %v", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("decoding config: %v", err)
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid configuration data: %v", err)
	}

	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load node's key: %v", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, ccfg.DefaultLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}

	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app), // TODO: NewUnsyncedLocalClientCreator(app) is other option which doesn't take a lock on all the connections to the app
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger)

	if err != nil {
		return nil, fmt.Errorf("creating node: %v", err)
	}

	return node, nil
}

func buildChainClient(cfg *config.KwildConfig, logger log.Logger) (chainClient.ChainClient, error) {
	fmt.Println("Building chain client", cfg.Deposits.ClientChainRPCURL, cfg.Deposits.ChainCode)
	return ccService.NewChainClient(cfg.Deposits.ClientChainRPCURL,
		ccService.WithChainCode(chainTypes.ChainCode(cfg.Deposits.ChainCode)),
		ccService.WithLogger(*logger.Named("chainClient")),
		ccService.WithReconnectInterval(int64(cfg.Deposits.ReconnectionInterval)),
		ccService.WithRequiredConfirmations(int64(cfg.Deposits.BlockConfirmations)),
	)
}

func buildAccountRepository(logger log.Logger, cfg *config.KwildConfig) (AccountStore, error) {
	if cfg.WithoutAccountStore {
		return balances.NewEmptyAccountStore(*logger.Named("emptyAccountStore")), nil
	}

	return balances.NewAccountStore(
		balances.WithLogger(*logger.Named("accountStore")),
		balances.WithPath(cfg.SqliteFilePath),
	)
}

type AccountStore interface {
	BatchCredit(creditList []*balances.Credit, chain *balances.ChainConfig) error
	BatchSpend(spendList []*balances.Spend, chain *balances.ChainConfig) error
	ChainExists(chainCode int32) (bool, error)
	Close() error
	CreateChain(chainCode int32, height int64) error
	Credit(credit *balances.Credit) error
	GetAccount(address string) (*balances.Account, error)
	GetHeight(chainCode int32) (int64, error)
	SetHeight(chainCode int32, height int64) error
	Spend(spend *balances.Spend) error
}

func buildChainSyncer(cfg *config.KwildConfig, cc chainClient.ChainClient, as AccountStore, logger log.Logger) (starter, error) {
	if cfg.WithoutChainSyncer {
		return chainsyncer.NewEmptyChainSyncer(), nil
	}

	walletAddress := kwilCrypto.AddressFromPrivateKey(cfg.PrivateKey)

	return chainsyncer.Builder().
		WithLogger(*logger.Named("chainSyncer")).
		WritesTo(as).
		ListensTo(cfg.Deposits.PoolAddress).
		WithChainClient(cc).
		WithReceiverAddress(walletAddress).
		WithChunkSize(int64(cfg.ChainSyncer.ChunkSize)).
		Build()
}

func buildTxSvc(ctx context.Context, cfg *config.KwildConfig, as AccountStore, logger log.Logger) (*txsvc.Service, error) {
	return txsvc.NewService(ctx, cfg,
		txsvc.WithLogger(*logger.Named("txService")),
		txsvc.WithAccountStore(as),
		txsvc.WithSqliteFilePath(cfg.SqliteFilePath),
		txsvc.WithExtensions(cfg.ExtensionEndpoints...),
	)
}

func buildHealthSvc(logger log.Logger) *healthsvc.Server {
	// health service
	registrar := healthcheck.NewRegistrar(*logger.Named("healthcheck"))
	registrar.RegisterAsyncCheck(10*time.Second, 3*time.Second, healthcheck.Check{
		Name: "dummy",
		Check: func(ctx context.Context) error {
			// error make this check fail, nil will make it succeed
			return nil
		},
	})
	ck := registrar.BuildChecker(simple_checker.New(logger))
	return healthsvc.NewServer(ck)
}

type starter interface {
	Start(ctx context.Context) error
}

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the kwild process",
		RunE: func(cmd *cobra.Command, args []string) error {
			syscall.Kill(1, syscall.SIGTERM)
			fmt.Printf("stopping kwild process\n")
			return nil
		},
	}
}
