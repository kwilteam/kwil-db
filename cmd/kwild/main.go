package main

import (
	"context"
	"fmt"
	"kwil/x/async"
	"kwil/x/sqlx/cache"
	"kwil/x/sqlx/env"
	"kwil/x/sqlx/manager"
	"kwil/x/sqlx/sqlclient"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kwil/x/cfgx"
	"kwil/x/grpcx"
	"kwil/x/logx"
	"kwil/x/proto/apipb"
	"kwil/x/proto/depositsvc"
	"kwil/x/service/apisvc"

	kg "kwil/cmd/kwild-gateway/server"
	deposits "kwil/x/deposits/app"

	"github.com/oklog/run"
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

	deposits, err := buildDeposits(cfg, client, chainClient, "274194b20d248d47c05913c039c65783647e527aa6360e5e143417f8bb50b988")
	if err != nil {
		return fmt.Errorf("failed to build deposits: %w", err)
	}

	mngrCfg := cfgx.GetConfig().Select("manager-settings")
	cache := cache.New()
	mngr, err := manager.New(ctx, client, mngrCfg, cache)
	if err != nil {
		return fmt.Errorf("failed to initialize new manager: %w", err)
	}
	err = mngr.SyncCache(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync cache: %w", err)
	}

	apiService := apisvc.NewService(mngr)
	httpHandler := apisvc.NewHandler(logger)

	return serve(ctx, logger, deposits, httpHandler, apiService)
}

func serve(ctx context.Context, logger logx.Logger, d *deposits.Service, httpHandler http.Handler, apiService apipb.KwilServiceServer) error {
	var g run.Group

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	g.Add(func() error {
		err = d.Sync(ctx)
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
		apipb.RegisterKwilServiceServer(grpcServer, apiService)
		depositsvc.RegisterKwilServiceServer(grpcServer, d)
		return grpcServer.Serve(listener)
	}, func(error) {
		_ = listener.Close()
	})

	httpServer := http.Server{
		Addr:    ":8081",
		Handler: httpHandler,
	}
	g.Add(func() error {
		return httpServer.ListenAndServe()
	}, func(error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
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
