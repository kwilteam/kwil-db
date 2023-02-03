package server

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/health/grpc_health_v1"
	infopb "kwil/api/protobuf/info/v0/gen/go"
	pricingpb "kwil/api/protobuf/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/internal/app/kwild/config"
	"kwil/internal/controller/grpc/v0/healthsvc"
	"kwil/internal/controller/grpc/v0/infosvc"
	"kwil/internal/controller/grpc/v0/pricingsvc"
	"kwil/internal/controller/grpc/v0/txsvc"
	"kwil/pkg/grpc/server"
	"kwil/pkg/logger"
	"kwil/x/deposits"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	cfg        config.ServerConfig
	logger     logger.Logger
	txSvc      *txsvc.Service
	infoSvc    *infosvc.Service
	pricingSvc *pricingsvc.Service
	healthSvc  *healthsvc.Server
	depositer  deposits.Depositer
}

func New(cfg config.ServerConfig, txSvc *txsvc.Service, infoSvc *infosvc.Service, pricingSvc *pricingsvc.Service,
	healthSvc *healthsvc.Server, depositer deposits.Depositer, logger logger.Logger) *Server {
	return &Server{
		cfg:        cfg,
		logger:     logger,
		txSvc:      txSvc,
		infoSvc:    infoSvc,
		pricingSvc: pricingSvc,
		healthSvc:  healthSvc,
		depositer:  depositer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting server")

	cancelCtx, done := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(cancelCtx)

	listener, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	g.Go(func() error {
		err = s.depositer.Start(gctx)
		if err != nil {
			return err
		}
		s.logger.Info("deposits synced")

		<-gctx.Done()
		return nil
	})

	g.Go(func() error {
		grpcServer := server.New(s.logger)
		txpb.RegisterTxServiceServer(grpcServer, s.txSvc)
		infopb.RegisterInfoServiceServer(grpcServer, s.infoSvc)
		pricingpb.RegisterPricingServiceServer(grpcServer, s.pricingSvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, s.healthSvc)
		s.logger.Info("grpc server started", zap.String("address", listener.Addr().String()))

		go func() {
			<-gctx.Done()
			grpcServer.Stop()
			s.logger.Info("grpc server stopped")
		}()

		return grpcServer.Serve(ctx, "0.0.0.0:50051")
	})

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-gctx.Done():
			s.logger.Info("close signal goroutine", zap.Error(gctx.Err()))
			return gctx.Err()
		case sig := <-c:
			s.logger.Warn("received signal", zap.String("signal", sig.String()))
			done()
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.logger.Info("server context is canceled")
			return nil
		} else {
			s.logger.Error("server error", zap.Error(err))
		}
	}

	return nil
}

//func (s *Server) Stop() error {
//	s.logger.Info("stopping server")
//	return nil
//}
