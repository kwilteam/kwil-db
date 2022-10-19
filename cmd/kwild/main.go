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

	"kwil/pkg/types/chain/pricing"
	"kwil/x/cfgx"
	"kwil/x/chain/auth"
	"kwil/x/chain/config"
	"kwil/x/chain/contracts"
	"kwil/x/chain/utils"
	"kwil/x/crypto"
	"kwil/x/deposits"
	"kwil/x/grpcx"
	"kwil/x/logx"
	"kwil/x/proto/apipb"
	"kwil/x/service/apisvc"

	"github.com/oklog/run"
)

func execute(logger logx.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Config
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(path)
	conf, err := config.LoadConfig("kwil_config.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	err = os.Setenv(cfgx.Meta_Config_Env, "meta-config.yaml")
	if err != nil {
		panic(err)
	}

	client, err := config.ConnectChain()
	if err != nil {
		return fmt.Errorf("failed to connect to client chain: %w", err)
	}

	dc := cfgx.GetConfig().Select("deposit-settings")

	d, err := deposits.New(dc)
	if err != nil {
		return fmt.Errorf("failed to initialize new deposits: %w", err)
	}
	// when should i defer close?  This function returns after ~1 second
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

	serv := apisvc.NewService(d, p, c)
	httpHandler := apisvc.NewHandler(logger, a.Authenticator)

	return serve(logger, httpHandler, serv)
}

func serve(logger logx.Logger, httpHandler http.Handler, srv apipb.KwilServiceServer) error {
	var g run.Group

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	g.Add(func() error {
		grpcServer := grpcx.NewServer(logger)
		apipb.RegisterKwilServiceServer(grpcServer, srv)
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
