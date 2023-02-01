package server

import (
	"context"
	"fmt"
	"github.com/oklog/run"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
	accountpb "kwil/api/protobuf/account/v0/gen/go"
	pricingpb "kwil/api/protobuf/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	hasura2 "kwil/internal/pkg/graphql/hasura"
	"kwil/internal/pkg/healthcheck"
	"kwil/internal/pkg/healthcheck/simple-checker"
	"kwil/pkg/logger"
	"kwil/pkg/sql/sqlclient"

	"kwil/internal/controller/grpc/v0/accountsvc"
	"kwil/internal/controller/grpc/v0/healthsvc"
	"kwil/internal/controller/grpc/v0/pricingsvc"
	"kwil/internal/controller/grpc/v0/txsvc"
	"kwil/kwil/repository"
	"kwil/pkg/grpc/server"
	"kwil/x/cfgx"
	"kwil/x/deposits"
	"kwil/x/execution/executor"
	"kwil/x/sqlx/env"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func execute(logger logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := sqlclient.Open(env.GetDbConnectionString(), 60*time.Second)
	if err != nil {
		return fmt.Errorf("failed to open sql client: %w", err)
	}

	cfg := cfgx.GetConfig()
	chainClient, err := buildChainClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to build chain client: %w", err)
	}

	// build repository prepared statement
	queries, err := repository.Prepare(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to prepare queries: %w", err)
	}

	deposits, err := buildDeposits(cfg, client, queries, chainClient)
	if err != nil {
		return fmt.Errorf("failed to build deposits: %w", err)
	}

	hasuraManager := hasura2.NewClient(viper.GetString(hasura2.GraphqlEndpointFlag))

	// build executor
	exec, err := executor.NewExecutor(ctx, client, queries, hasuraManager)
	if err != nil {
		return fmt.Errorf("failed to build executor: %w", err)
	}

	// build account service
	accountService := accountsvc.NewService(queries)

	// pricing service
	pricingService := pricingsvc.NewService()

	// tx service
	txService := txsvc.NewService(queries, exec)

	// health service
	registrar := healthcheck.NewRegistrar()
	registrar.RegisterAsyncCheck(10*time.Second, 15*time.Second, healthcheck.Check{
		Name: "dummy",
		Check: func(ctx context.Context) error {
			// error make this check fail, nil will make it succeed
			return nil
		},
	})
	ck := registrar.BuildChecker(simple_checker.New())
	healthService := healthsvc.NewServer(ck)

	return serve(ctx, logger, txService, accountService, pricingService, healthService, deposits)
}

func serve(ctx context.Context, logger logger.Logger, txSvc *txsvc.Service, accountSvc *accountsvc.Service, pricingSvc *pricingsvc.Service, healthSvc *healthsvc.Server, depsts deposits.Depositer) error {
	var g run.Group

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	g.Add(func() error {
		err = depsts.Start(ctx)
		if err != nil {
			return err
		}
		logger.Info("deposits synced")

		<-ctx.Done() // if any rungroup actor returns, the whole group stops, so we wait for ctx.Done() to return
		return nil
	}, func(err error) {
		if err != nil {
			logger.Sugar().Errorf("deposits failed to sync: %s", err)
		}
	})

	g.Add(func() error {
		grpcServer := server.New(logger)
		txpb.RegisterTxServiceServer(grpcServer, txSvc)
		accountpb.RegisterAccountServiceServer(grpcServer, accountSvc)
		pricingpb.RegisterPricingServiceServer(grpcServer, pricingSvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthSvc)
		logger.Info("grpc server started", zap.String("address", listener.Addr().String()))
		return grpcServer.Serve(ctx, "0.0.0.0:50051")
	}, func(error) {
		logger.Error("grpc server stopped", zap.Error(err))
	})

	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
			return nil
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})

	return g.Run()
}

func isGatewayEnabled() bool {
	var args []string
	with_gateway_flag := false
	found := -2
	for i, arg := range os.Args {
		if i == found+1 {
			if arg == "true" {
				with_gateway_flag = true
			}
			continue
		}

		if arg != "--withgateway" {
			args = append(args, arg)
			continue
		}

		found = i
	}

	if with_gateway_flag {
		os.Args = args //make sure the flag and value are removed
	} else {
		with_gateway_env := os.Getenv("WITH_GATEWAY")
		with_gateway_flag = with_gateway_env == "true"
	}

	return with_gateway_flag
}
