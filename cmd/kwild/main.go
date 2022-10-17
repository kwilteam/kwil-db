package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/oklog/run"
	"kwil/pkg/types/chain/pricing"
	"kwil/x/api/handler"
	"kwil/x/api/service"
	v0 "kwil/x/api/v0"
	"kwil/x/chain/auth"
	"kwil/x/chain/config"
	"kwil/x/chain/contracts"
	"kwil/x/chain/utils"
	nc "kwil/x/common/config"
	"kwil/x/crypto"
	nd "kwil/x/deposits"
	"kwil/x/grpcx"
	"kwil/x/logx"
)

const (
	grpcPortEnv = "KWIL_GRPC_PORT"
	httpPortEnv = "KWIL_HTTP_PORT"
)

func execute(logger logx.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Config
	conf, err := config.LoadConfig("kwil_config.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := config.ConnectChain()
	if err != nil {
		return fmt.Errorf("failed to connect to client chain: %w", err)
	}

	cb := nc.Builder()
	depConf, err := cb.UseFile("deposit-config.yaml").Build()
	if err != nil {
		return fmt.Errorf("failed to load deposit config: %w", err)
	}

	d, err := nd.New(depConf)
	defer d.Close()
	if err != nil {
		return fmt.Errorf("failed to initialize new deposits: %w", err)
	}
	d.Listen(ctx)

	kr, err := crypto.NewKeyring(conf)
	if err != nil {
		return fmt.Errorf("failed to create keyring: %w", err)
	}

	acc, err := kr.GetDefaultAccount()
	if err != nil {
		return fmt.Errorf("failed to get default account: %w", err)
	}
	a := auth.NewAuth(conf, acc)
	a.Client.AuthAll()

	ppath := conf.GetPricePath()
	pbytes, err := utils.LoadFileFromRoot(ppath)
	if err != nil {
		return fmt.Errorf("failed to load pricing config: %w", err)
	}

	p, err := pricing.New(pbytes)
	if err != nil {
		return fmt.Errorf("failed to initialize pricing: %w", err)
	}

	c, err := contracts.NewContractClient(acc, client, conf.ClientChain.DepositContract.Address, strconv.Itoa(conf.ClientChain.ChainID))
	if err != nil {
		return fmt.Errorf("failed to initialize contract client: %w", err)
	}

	serv := service.NewService(d, p, c)
	httpHandler := handler.New(logger, a.Authenticator)

	return serve(logger, httpHandler, serv)
}

func serve(logger logx.Logger, httpHandler http.Handler, srv v0.KwilServiceServer) error {
	var g run.Group

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	g.Add(func() error {
		grpcServer := grpcx.NewServer(logger)
		v0.RegisterKwilServiceServer(grpcServer, srv)
		return grpcServer.Serve(listener)
	}, func(error) {
		listener.Close()
	})

	httpServer := http.Server{
		Addr:    ":8080",
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

	if err := execute(logger); err != nil {
		logger.Sugar().Error(err)
	}
}
