package server

import (
	"context"
	"errors"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/internal/app/kwild/config"
	"kwil/internal/controller/grpc/healthsvc/v0"
	"kwil/internal/controller/grpc/txsvc/v1"
	chainsyncer "kwil/pkg/balances/chain-syncer"
	"kwil/pkg/grpc/server"
	"kwil/pkg/log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	Cfg         *config.KwildConfig
	Log         log.Logger
	HealthSvc   *healthsvc.Server
	ChainSyncer *chainsyncer.ChainSyncer
	TxSvc       *txsvc.Service
}

func (s *Server) Start(ctx context.Context) error {
	s.Log.Info("starting server")

	cancelCtx, done := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(cancelCtx)

	g.Go(func() error {
		if err := s.ChainSyncer.Start(gctx); err != nil {
			return err
		}
		s.Log.Info("deposits synced")

		<-gctx.Done()
		return nil
	})

	g.Go(func() error {
		grpcServer := server.New(s.Log)
		txpb.RegisterTxServiceServer(grpcServer, s.TxSvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, s.HealthSvc)

		go func() {
			<-gctx.Done()
			grpcServer.Stop()
			s.Log.Info("grpc server stopped")
		}()

		return grpcServer.Serve(ctx, s.Cfg.GrpcListenAddress)
	})
	s.Log.Info("grpc server started", zap.String("address", s.Cfg.GrpcListenAddress))

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-gctx.Done():
			s.Log.Info("close signal goroutine", zap.Error(gctx.Err()))
			return gctx.Err()
		case sig := <-c:
			s.Log.Warn("received signal", zap.String("signal", sig.String()))
			done()
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.Log.Info("server context is canceled")
			return nil
		} else {
			s.Log.Error("server error", zap.Error(err))
		}
	}

	return nil
}

//func (s *Server) Stop() error {
//	s.log.Info("stopping server")
//	return nil
//}
