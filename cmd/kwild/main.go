package main

import (
	"context"
	"fmt"
	"kwil/x/deposits/processor"
	"kwil/x/svcx/wallet"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kwil/pkg/types/chain/pricing"
	"kwil/x/cfgx"
	"kwil/x/crypto"
	"kwil/x/deposits"
	"kwil/x/grpcx"
	"kwil/x/logx"
	"kwil/x/proto/apipb"
	"kwil/x/service/apisvc"
	"kwil/x/utils"

	"github.com/oklog/run"
)

func execute(logger logx.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc := cfgx.GetConfig().Select("deposit-settings")

	kr, err := crypto.NewKeyring(dc)
	if err != nil {
		return fmt.Errorf("failed to create keyring: %w", err)
	}

	acc, err := kr.GetDefaultAccount()
	if err != nil {
		return fmt.Errorf("failed to get default account: %w", err)
	}

	wrs, err := loadWalletService(logger)
	if err != nil {
		return fmt.Errorf("failed to load wallet service: %w", err)
	}

	d, err := deposits.New(dc, logger, acc)
	if err != nil {
		return fmt.Errorf("failed to initialize new deposits: %w", err)
	}
	d.Listen(ctx)

	ppath := "./prices.json"
	pbytes, err := utils.LoadFileFromRoot(ppath)
	if err != nil {
		return fmt.Errorf("failed to load pricing config: %w", err)
	}

	p, err := pricing.New(pbytes)
	if err != nil {
		return fmt.Errorf("failed to initialize pricing: %w", err)
	}

	serv := apisvc.NewService(d, p, wrs)
	httpHandler := apisvc.NewHandler(logger)

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

func loadWalletService(l logx.Logger) (wallet.RequestService, error) {
	tr := processor.AsMessageTransform(processor.NewProcessor(l))

	p, err := wallet.NewRequestProcessor(cfgx.GetConfig(), tr)
	if err != nil {
		return nil, err
	}

	w, err := wallet.NewRequestService(cfgx.GetConfig())
	if err != nil {
		return nil, err
	}

	err = p.Start()
	if err != nil {
		return nil, err
	}

	err = w.Start()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func main() {
	logger := logx.New()

	if err := execute(logger); err != nil {
		logger.Sugar().Error(err)
	}
}
