package server

import (
	"context"
	"errors"
	"kwil/internal/app/kwild/config"
	chainsyncer "kwil/pkg/balances/chain-syncer"
	grpcServer "kwil/pkg/grpc/server"
	"kwil/pkg/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	Ctx         context.Context
	Cfg         *config.KwildConfig
	Log         log.Logger
	ChainSyncer *chainsyncer.ChainSyncer
	Http        *GWServer
	Grpc        *grpcServer.Server

	done context.CancelFunc
}

func (s *Server) Start(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			s.Log.Error("kwild server panic", zap.Any("error", err))
		}
	}()

	s.Log.Info("starting server...")
	s.Log.Info("using new retry version")

	// graceful shutdown when receive signal
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	cancelCtx, done := context.WithCancel(ctx)
	s.done = done
	g, gctx := errgroup.WithContext(cancelCtx)

	g.Go(func() error {
		go func() {
			<-gctx.Done()
			s.Log.Info("stop http server")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer func() {
				cancel()
			}()
			if err := s.Http.Shutdown(ctx); err != nil {
				s.Log.Error("http server shutdown error", zap.Error(err))
			}
		}()
		return s.Http.Serve()
	})
	s.Log.Info("http server started", zap.String("address", s.Cfg.HttpListenAddress))

	g.Go(func() error {
		if err := s.ChainSyncer.Start(gctx); err != nil {
			return err
		}
		s.Log.Info("deposits synced")
		return nil
	})

	g.Go(func() error {
		go func() {
			<-gctx.Done()
			s.Log.Info("stop grpc server")
			s.Grpc.Stop()
		}()

		return s.Grpc.Serve(ctx, s.Cfg.GrpcListenAddress)
	})
	s.Log.Info("grpc server started", zap.String("address", s.Cfg.GrpcListenAddress))

	g.Go(func() error {
		select {
		case <-gctx.Done():
			s.Log.Info("close signal goroutine", zap.Error(gctx.Err()))
			return gctx.Err()
		case sig := <-gracefulShutdown:
			s.Log.Warn("received signal", zap.String("signal", sig.String()))
			s.Stop()
		}
		return nil
	})
	s.Log.Info("signal watcher started")

	err := g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.Log.Info("server context is canceled")
			return nil
		} else if errors.Is(err, http.ErrServerClosed) {
			s.Log.Info("http server is closed")
		} else {
			s.Log.Error("server error", zap.Error(err))
		}
	}

	return nil
}

func (s *Server) Stop() {
	s.Log.Warn("stop kwild services")
	s.done()
}
