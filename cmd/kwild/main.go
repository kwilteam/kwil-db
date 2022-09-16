package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kwilteam/kwil-db/internal/api/handler"
	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/kwilteam/kwil-db/internal/api/service"
	"github.com/kwilteam/kwil-db/internal/chain/auth"
	"github.com/kwilteam/kwil-db/internal/chain/config"
	"github.com/kwilteam/kwil-db/internal/chain/crypto"
	"github.com/kwilteam/kwil-db/internal/chain/deposits"
	"github.com/kwilteam/kwil-db/internal/chain/utils"
	"github.com/kwilteam/kwil-db/internal/common/logging"
	"github.com/kwilteam/kwil-db/pkg/types/chain/pricing"
	"github.com/oklog/run"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	grpcPortEnv = "KWIL_GRPC_PORT"
	httpPortEnv = "KWIL_HTTP_PORT"
)

func serve() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := config.InitBuildInfo()
	if err != nil {
		return fmt.Errorf("failed to initialize build info: %w", err)
	}

	// Load Config
	conf, err := config.LoadConfig("kwil_config.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize the global logger
	logging.InitLogger(config.BuildInfo.Version, conf.Log.Debug, conf.Log.Human)

	client, err := config.ConnectChain()
	if err != nil {
		return fmt.Errorf("failed to connect to client chain: %w", err)
	}

	d, err := deposits.Init(ctx, conf, client)
	if err != nil {
		return fmt.Errorf("failed to initialize deposits: %w", err)
	}

	defer d.Store.Close()

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

	fmt.Println("Node is running properly!")

	var g run.Group
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	g.Add(func() error {
		grpcServer := grpc.NewServer()
		serv := service.NewService(d.Store, p)
		v0.RegisterKwilServiceServer(grpcServer, serv)
		return grpcServer.Serve(listener)
	}, func(error) {
		listener.Close()
	})

	ath := auth.NewAuth(conf, acc)
	httpHandler := handler.NewHandler(ath.Authenticator)

	g.Add(func() error {
		return httpHandler.Server.ListenAndServe()
	}, func(error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpHandler.Server.Shutdown(ctx)
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
	if err := serve(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
