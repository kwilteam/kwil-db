package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"kwil/internal/app/kwild/config"
	"kwil/internal/app/kwild/server"
	"kwil/internal/controller/grpc/v0/accountsvc"
	"kwil/internal/controller/grpc/v0/healthsvc"
	"kwil/internal/controller/grpc/v0/pricingsvc"
	"kwil/internal/controller/grpc/v0/txsvc"
	"kwil/internal/pkg/graphql/hasura"
	"kwil/internal/pkg/healthcheck"
	simple_checker "kwil/internal/pkg/healthcheck/simple-checker"
	"kwil/kwil/repository"
	"kwil/pkg/chain/client/service"
	"kwil/pkg/log"
	"kwil/pkg/sql/sqlclient"
	"kwil/x/deposits"
	"kwil/x/execution/executor"
	"path/filepath"
	"time"
)

var RootCmd = &cobra.Command{
	Use:   "kwild",
	Short: "kwil grpc server",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		client, err := sqlclient.Open(cfg.Db.DbUrl(), 60*time.Second)
		if err != nil {
			return fmt.Errorf("failed to open sql client: %w", err)
		}

		chainClient, err := service.NewChainClientExplicit(&cfg.Fund.Chain, logger)
		if err != nil {
			return fmt.Errorf("failed to build chain client: %w", err)
		}

		// build repository prepared statement
		queries, err := repository.Prepare(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to prepare queries: %w", err)
		}

		dps, err := deposits.NewDepositer(&cfg.Fund, client, queries, chainClient, cfg.Fund.Wallet, logger)
		if err != nil {
			return fmt.Errorf("failed to build deposits: %w", err)
		}

		hasuraManager := hasura.NewClient(cfg.Graphql.Endpoint)

		// build executor
		exec, err := executor.NewExecutor(ctx, client, queries, hasuraManager, logger)
		if err != nil {
			return fmt.Errorf("failed to build executor: %w", err)
		}

		// build config service
		accountService := accountsvc.NewService(queries, logger)

		// pricing service
		pricingService := pricingsvc.NewService()

		// tx service
		txService := txsvc.NewService(queries, exec, logger)

		// health service
		registrar := healthcheck.NewRegistrar(logger)
		registrar.RegisterAsyncCheck(10*time.Second, 15*time.Second, healthcheck.Check{
			Name: "dummy",
			Check: func(ctx context.Context) error {
				// error make this check fail, nil will make it succeed
				return nil
			},
		})
		ck := registrar.BuildChecker(simple_checker.New(logger))
		healthService := healthsvc.NewServer(ck)

		// build server
		svr := server.New(cfg.Server, txService, accountService, pricingService, healthService, dps, logger)
		return svr.Start(ctx)
	}}

func init() {
	defaultConfigPath := filepath.Join("$HOME", config.DefaultConfigDir,
		fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigType))
	RootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file to use (default: '%s')", defaultConfigPath))

	config.BindGlobalFlags(RootCmd.PersistentFlags())
	config.BindGlobalEnv(RootCmd.PersistentFlags())
}
