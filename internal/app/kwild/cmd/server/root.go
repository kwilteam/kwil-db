package server

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
	"github.com/kwilteam/kwil-db/pkg/arweave"
	grpcServer "github.com/kwilteam/kwil-db/pkg/grpc/server"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"go.uber.org/zap"

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
	Short: "kwil grpc server",
	Long:  "Starts node with Kwild and CometBFT services",
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
	opts := []txsvc.TxSvcOpt{
		txsvc.WithLogger(*logger.Named("txService")),
		txsvc.WithAccountStore(as),
		txsvc.WithSqliteFilePath(cfg.SqliteFilePath),
		txsvc.WithExtensions(cfg.ExtensionEndpoints...),
	}

	if cfg.ArweaveConfig.BundlrURL != "" && cfg.PrivateKey != nil {
		opts = append(opts, buildArweaveWriter(cfg, logger))
	}

	return txsvc.NewService(ctx, cfg, opts...)
}

func buildArweaveWriter(cfg *config.KwildConfig, logger log.Logger) txsvc.TxSvcOpt {
	bundlrClient, err := arweave.NewBundlrClient(cfg.ArweaveConfig.BundlrURL, cfg.PrivateKey)
	if err != nil {
		panic("failed to create arweave bundlr client")
	}

	return txsvc.WithTxHook(func(tx *kTx.Transaction) error {
		bts, err := tx.Bytes()
		if err != nil {
			return err
		}

		res, err := bundlrClient.StoreItem(bts)
		if err != nil {
			logger.Error("failed to store tx on bundlr", zap.Error(err))
			return nil
		}

		logger.Info("tx stored on bundlr", zap.String("txId", res.TxID))

		return nil
	})
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
