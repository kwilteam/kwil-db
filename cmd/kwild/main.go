package main

import (
	"context"
	"fmt"
	"kwil/kwil/repository"
	"kwil/kwil/svc/accountsvc"
	"kwil/kwil/svc/pricingsvc"
	"kwil/kwil/svc/txsvc"
	"kwil/x/async"
	"kwil/x/deposits"
	"kwil/x/execution/executor"
	"kwil/x/graphql/hasura"
	"kwil/x/grpcx"
	"kwil/x/proto/accountspb"
	"kwil/x/proto/pricingpb"
	"kwil/x/proto/txpb"
	"kwil/x/sqlx/env"
	"kwil/x/sqlx/sqlclient"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kwil/x/cfgx"
	"kwil/x/logx"

	kg "kwil/cmd/kwil-gateway/server"

	"github.com/oklog/run"
	"github.com/spf13/viper"
)

func execute(logger logx.Logger) error {
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

	deposits, err := buildDeposits(cfg, client, queries, chainClient, "274194b20d248d47c05913c039c65783647e527aa6360e5e143417f8bb50b988")
	if err != nil {
		return fmt.Errorf("failed to build deposits: %w", err)
	}

	hasuraManager := hasura.NewClient(viper.GetString(hasura.GraphqlEndpointName))

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

	return serve(ctx, logger, txService, accountService, pricingService, deposits)
}

func serve(ctx context.Context, logger logx.Logger, txSvc *txsvc.Service, accountSvc *accountsvc.Service, pricingSvc *pricingsvc.Service, depsts deposits.Depositer) error {
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
			logger.Sugar().Errorf("deposits failed to sync: %d", err)
		}
	})

	g.Add(func() error {

		grpcServer := grpcx.NewServer(logger)
		txpb.RegisterTxServiceServer(grpcServer, txSvc)
		accountspb.RegisterAccountServiceServer(grpcServer, accountSvc)
		pricingpb.RegisterPricingServiceServer(grpcServer, pricingSvc)
		return grpcServer.Serve(listener)

	}, func(error) {
		_ = listener.Close()
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

func main() {
	logger := logx.New()

	stop := func(err error) {
		logger.Sugar().Error(err)
		os.Exit(1)
	}

	kwild := func() error {
		return execute(logger)
	}

	if !isGatewayEnabled() {
		if err := kwild(); err != nil {
			stop(err)
		}
	}

	async.Run(kg.Start).Catch(stop)

	<-async.Run(kwild).Catch(stop).DoneCh()
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
