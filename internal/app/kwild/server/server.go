package server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/health/grpc_health_v1"
	accountpb "kwil/api/protobuf/kwil/account/v0/gen/go"
	cfgpb "kwil/api/protobuf/kwil/configuration/v0/gen/go"
	pricingpb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/internal/app/kwild/config"
	"kwil/internal/controller/grpc/v0/accountsvc"
	"kwil/internal/controller/grpc/v0/configsvc"
	"kwil/internal/controller/grpc/v0/healthsvc"
	"kwil/internal/controller/grpc/v0/pricingsvc"
	"kwil/internal/controller/grpc/v0/txsvc"
	"kwil/internal/pkg/deposits"
	"kwil/pkg/grpc/server"
	"kwil/pkg/log"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	cfg        config.ServerConfig
	logger     log.Logger
	txSvc      *txsvc.Service
	accountSvc *accountsvc.Service
	configsvc  *configsvc.Service
	pricingSvc *pricingsvc.Service
	healthSvc  *healthsvc.Server
	depositer  deposits.Depositer
}

func New(cfg config.ServerConfig, txSvc *txsvc.Service, accSvc *accountsvc.Service, cfgSvc *configsvc.Service,
	healthSvc *healthsvc.Server, prcSvc *pricingsvc.Service, depositer deposits.Depositer, logger log.Logger) *Server {
	return &Server{
		cfg:        cfg,
		logger:     logger,
		txSvc:      txSvc,
		accountSvc: accSvc,
		configsvc:  cfgSvc,
		pricingSvc: prcSvc,
		healthSvc:  healthSvc,
		depositer:  depositer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting server")

	cancelCtx, done := context.WithCancel(ctx)
	g, gctx := errgroup.WithContext(cancelCtx)

	g.Go(func() error {
		if err := s.depositer.Start(gctx); err != nil {
			return err
		}
		s.logger.Info("deposits synced")

		<-gctx.Done()
		return nil
	})

	g.Go(func() error {
		grpcServer := server.New(s.logger)
		txpb.RegisterTxServiceServer(grpcServer, s.txSvc)
		accountpb.RegisterAccountServiceServer(grpcServer, s.accountSvc)
		cfgpb.RegisterConfigServiceServer(grpcServer, s.configsvc)
		pricingpb.RegisterPricingServiceServer(grpcServer, s.pricingSvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, s.healthSvc)

		go func() {
			<-gctx.Done()
			grpcServer.Stop()
			s.logger.Info("grpc server stopped")
		}()

		return grpcServer.Serve(ctx, s.cfg.Addr)
	})
	s.logger.Info("grpc server started", zap.String("address", s.cfg.Addr))

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

	err := g.Wait()
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
//	s.log.Info("stopping server")
//	return nil
//}
