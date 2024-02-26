package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/abci"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft"
	"github.com/kwilteam/kwil-db/pkg/grpc/gateway"
	grpc "github.com/kwilteam/kwil-db/pkg/grpc/server"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Server controls the gRPC server and http gateway.
type Server struct {
	grpcServer   *grpc.Server
	gateway      *gateway.GatewayServer
	cometBftNode *cometbft.CometBftNode
	closers      *closeFuncs
	log          log.Logger

	cfg *config.KwildConfig

	cancelCtxFunc context.CancelFunc
}

func (s *Server) Start(ctx context.Context) error {
	defer func() {
		err := s.closers.closeAll()
		if err != nil {
			s.log.Error("failed to close resource:", zap.Error(err))
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			switch et := err.(type) {
			case abci.FatalError:
				s.log.Error("Blockchain application hit an unrecoverable error:\n\n%v",
					zap.Stringer("error", et))
				// cometbft *may* already recover panics from the application. Investigate.
			default:
				s.log.Error("kwild server panic", zap.Any("error", err))
			}
		}
	}()

	s.log.Info("starting server...")

	cancelCtx, done := context.WithCancel(ctx)
	s.cancelCtxFunc = done

	group, groupCtx := errgroup.WithContext(cancelCtx)

	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop http server")
			ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.gateway.Shutdown(ctx2); err != nil {
				s.log.Error("http server shutdown error", zap.Error(err))
			}
		}()

		s.log.Info("http server started", zap.String("address", s.cfg.AppCfg.HttpListenAddress))
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
	s.log.Info("grpc server started", zap.String("address", s.cfg.AppCfg.GrpcListenAddress))

	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop comet server")
			if err := s.cometBftNode.Stop(); err != nil {
				s.log.Warn("failed to stop comet server", zap.Error(err))
			}
		}()

		return s.cometBftNode.Start()
	})
	s.log.Info("comet node started")

	err := group.Wait()

	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.log.Info("server context is canceled")
			return nil
		}
		if errors.Is(err, http.ErrServerClosed) {
			s.log.Info("http server is closed")
		} else {
			s.log.Error("server error", zap.Error(err))
			s.cancelCtxFunc()
			return err
		}
	}

	return nil
}

// Stop begins shutting down the Server. However, the caller of Start will
// normally cancel the provided context and wait for it to return.
func (s *Server) Stop() {
	s.log.Warn("stop kwild services")
	s.cancelCtxFunc()
}
