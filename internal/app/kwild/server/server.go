package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/grpc/gateway"
	grpc "github.com/kwilteam/kwil-db/pkg/grpc/server"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// comet bft node has a gross and confusing interface, so we use this to make it more clear
type StartStopper interface {
	Start() error
	Stop() error
}

type Server struct {
	grpcServer   *grpc.Server
	gateway      *gateway.GatewayServer
	cometBftNode StartStopper
	log          log.Logger

	cfg *config.KwildConfig

	cancelCtxFunc context.CancelFunc
}

func (s *Server) Start(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			s.log.Error("kwild server panic", zap.Any("error", err))
		}
	}()

	s.log.Info("starting server...")

	// graceful shutdown when receive signal
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	cancelCtx, done := context.WithCancel(ctx)
	s.cancelCtxFunc = done

	group, groupCtx := errgroup.WithContext(cancelCtx)

	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop http server")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer func() {
				cancel()
			}()
			if err := s.gateway.Shutdown(ctx); err != nil {
				s.log.Error("http server shutdown error", zap.Error(err))
			}
		}()

		s.log.Info("http server started", zap.String("address", s.cfg.HttpListenAddress))
		return s.gateway.Start()
	})

	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop grpc server")
			s.grpcServer.Stop()
		}()

		return s.grpcServer.Start()
	})
	s.log.Info("grpc server started", zap.String("address", s.cfg.GrpcListenAddress))

	group.Go(func() error {
		select {
		case <-groupCtx.Done():
			s.log.Info("close signal goroutine", zap.Error(groupCtx.Err()))
			return groupCtx.Err()
		case sig := <-gracefulShutdown:
			s.log.Warn("received signal", zap.String("signal", sig.String()))
			s.Stop()
		}
		return nil
	})
	s.log.Info("signal watcher started")

	err := group.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.log.Info("server context is canceled")
			return nil
		} else if errors.Is(err, http.ErrServerClosed) {
			s.log.Info("http server is closed")
		} else {
			s.log.Error("server error", zap.Error(err))
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.log.Warn("stop kwild services")
	s.cancelCtxFunc()
}
