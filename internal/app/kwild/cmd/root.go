package cmd

import (
	"context"
	"fmt"
	"syscall"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
	"github.com/kwilteam/kwil-db/internal/controller/grpc/healthsvc/v0"
	"github.com/kwilteam/kwil-db/internal/controller/grpc/txsvc/v1"
	"github.com/kwilteam/kwil-db/internal/pkg/gateway/middleware/cors"
	"github.com/kwilteam/kwil-db/internal/pkg/healthcheck"
	simple_checker "github.com/kwilteam/kwil-db/internal/pkg/healthcheck/simple-checker"
	grpcServer "github.com/kwilteam/kwil-db/pkg/grpc/server"

	"google.golang.org/grpc/health/grpc_health_v1"

	"time"

	"github.com/kwilteam/kwil-db/pkg/balances"
	chainsyncer "github.com/kwilteam/kwil-db/pkg/balances/chain-syncer"
	chainClient "github.com/kwilteam/kwil-db/pkg/chain/client"
	ccService "github.com/kwilteam/kwil-db/pkg/chain/client/service" // shorthand for chain client service
	chainTypes "github.com/kwilteam/kwil-db/pkg/chain/types"
	kwilCrypto "github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	return startCmd
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts kwild server",
	Long:  "Starts node with Kwild services",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := config.LoadKwildConfig()
		if err != nil {
			return err
		}

		logger := log.New(cfg.Log)
		logger = *logger.Named("kwild")

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		chainclient, err := buildChainClient(cfg, logger)
		if err != nil {
			return fmt.Errorf("failed to build chain client: %w", err)
		}

		accountStore, err := buildAccountRepository(logger, cfg)
		if err != nil {
			return fmt.Errorf("failed to build account repository: %w", err)
		}

		chainSyncer, err := buildChainSyncer(cfg, chainclient, accountStore, logger)
		if err != nil {
			return fmt.Errorf("failed to build chain syncer: %w", err)
		}

		txSvc, err := buildTxSvc(ctx, cfg, accountStore, logger)
		if err != nil {
			return fmt.Errorf("failed to build tx service: %w", err)
		}

		healthSvc := buildHealthSvc(logger)

		// kwil gateway
		gw := server.NewGWServer(runtime.NewServeMux(), cfg, logger)

		if err := gw.SetupGrpcSvc(ctx); err != nil {
			return err
		}
		if err := gw.SetupHTTPSvc(ctx); err != nil {
			return err
		}

		gw.AddMiddlewares(
			// from innermost middleware
			//auth.MAuth(keyManager, logger),
			cors.MCors([]string{}),
		)

		// grpc server
		grpcServer := grpcServer.New(logger)
		txpb.RegisterTxServiceServer(grpcServer, txSvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthSvc)

		// start
		server := &server.Server{
			Cfg:         cfg,
			Log:         logger,
			ChainSyncer: chainSyncer,
			Http:        gw,
			Grpc:        grpcServer,
		}

		return server.Start(ctx)
	}}

func init() {
	/*
		defaultConfigPath := filepath.Join("$HOME", config.DefaultConfigDir,
			fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigType))
		RootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file to use (default: '%s')", defaultConfigPath))
	*/

	config.BindFlagsAndEnv(startCmd.PersistentFlags())
}

func buildChainClient(cfg *config.KwildConfig, logger log.Logger) (chainClient.ChainClient, error) {
	return ccService.NewChainClient(cfg.Deposits.ClientChainRPCURL,
		ccService.WithChainCode(chainTypes.ChainCode(cfg.Deposits.ChainCode)),
		ccService.WithLogger(*logger.Named("chainClient")),
		ccService.WithReconnectInterval(int64(cfg.Deposits.ReconnectionInterval)),
		ccService.WithRequiredConfirmations(int64(cfg.Deposits.BlockConfirmations)),
	)
}

func buildAccountRepository(logger log.Logger, cfg *config.KwildConfig) (*balances.AccountStore, error) {
	return balances.NewAccountStore(
		balances.WithLogger(*logger.Named("accountStore")),
		balances.WithPath(cfg.SqliteFilePath),
	)
}

func buildChainSyncer(cfg *config.KwildConfig, cc chainClient.ChainClient, as *balances.AccountStore, logger log.Logger) (*chainsyncer.ChainSyncer, error) {
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

func buildTxSvc(ctx context.Context, cfg *config.KwildConfig, as *balances.AccountStore, logger log.Logger) (*txsvc.Service, error) {
	return txsvc.NewService(ctx, cfg,
		txsvc.WithLogger(*logger.Named("txService")),
		txsvc.WithAccountStore(as),
		txsvc.WithSqliteFilePath(cfg.SqliteFilePath),
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

// from v0, removed 04/03/23
/*
	ctx := cmd.Context()
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// build log
	//log, err := log.NewLogger(cfg.log)
	logger := log.New(cfg.Log)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := sqlclient.Open(cfg.DB.DbUrl(), 60*time.Second)
	if err != nil {
		return fmt.Errorf("failed to open sql client: %w", err)
	}

	//&cfg.Fund.Chain, logger
	chainClient, err := service.NewChainClient(cfg.Fund.Chain.RpcUrl,
		service.WithChainCode(chainTypes.ChainCode(cfg.Fund.Chain.ChainCode)),
		service.WithLogger(logger),
		service.WithReconnectInterval(cfg.Fund.Chain.ReconnectInterval),
		service.WithRequiredConfirmations(cfg.Fund.Chain.BlockConfirmation),
	)
	if err != nil {
		return fmt.Errorf("failed to build chain client: %w", err)
	}

	// build repository prepared statement
	queries, err := repository.Prepare(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to prepare queries: %w", err)
	}

	dps, err := deposits.NewDepositer(cfg.Fund.PoolAddress, client, queries, chainClient, cfg.Fund.Wallet, logger)
	if err != nil {
		return fmt.Errorf("failed to build deposits: %w", err)
	}

	hasuraManager := hasura.NewClient(cfg.Graphql.Addr, logger)
	go hasura.Initialize(cfg.Graphql.Addr, logger)

	// build executor
	exec, err := executor.NewExecutor(ctx, client, queries, hasuraManager, logger)
	if err != nil {
		return fmt.Errorf("failed to build executor: %w", err)
	}

	// build config service
	accSvc := accountsvc.NewService(queries, logger)

	// pricing service
	prcSvc := pricingsvc.NewService(exec)

	// tx service
	txService := txsvc.NewService(queries, exec, logger)

	// health service
	registrar := healthcheck.NewRegistrar(logger)
	registrar.RegisterAsyncCheck(10*time.Second, 3*time.Second, healthcheck.Check{
		Name: "dummy",
		Check: func(ctx context.Context) error {
			// error make this check fail, nil will make it succeed
			return nil
		},
	})
	ck := registrar.BuildChecker(simple_checker.New(logger))
	healthService := healthsvc.NewServer(ck)

	// configuration service
	cfgService := configsvc.NewService(cfg, logger)
	// build server
	svr := server.New(cfg.Server, txService, accSvc, cfgService, healthService, prcSvc, dps, logger)
	return svr.Start(ctx)
*/
