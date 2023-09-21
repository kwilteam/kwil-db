// Package server defines the main Kwil server, which includes the blockchain
// node and the gRPC services that interface with the Kwil dataset engine.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
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

const (
	// Top-level directory structure for the Server's systems
	abciDirName        = kwild.ABCIDirName
	applicationDirName = kwild.ApplicationDirName
	rcvdSnapsDirName   = kwild.ReceivedSnapsDirName
	signingDirName     = kwild.SigningDirName

	// Note that the sqlLite file path is user-configurable
	// e.g. "data/kwil.db"
)

// New builds the kwild server.
func New(ctx context.Context, cfg *config.KwildConfig) (svr *Server, err error) {
	closers := &closeFuncs{
		closers: make([]func() error, 0),
	}

	defer func() {
		if r := recover(); r != nil {
			svr = nil
			err = fmt.Errorf("panic while building kwild: %v", r)
			closers.closeAll()
		}
	}()

	logger, err := log.NewChecked(*cfg.LogConfig())
	if err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}
	logger = *logger.Named("kwild")

	dbDir := cfg.AppCfg.SqliteFilePath
	if dbDir == "" { // config parsing should have set it (sanitizeCfgPaths), but we can generate the default too
		dbDir = filepath.Join(cfg.RootDir, config.DefaultSQLitePath)
		logger.Warn("using fallback sqlite path", zap.String("sqlitePath", dbDir))
	}

	deps := &coreDependencies{
		ctx:    ctx,
		cfg:    cfg,
		log:    logger,
		opener: newSqliteOpener(dbDir),
	}

	return buildServer(deps, closers), nil
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
			s.log.Error("kwild server panic", zap.Any("error", err))
			/* xxx
			switch et := err.(type) {
			case abci.FatalError:
				s.log.Error("Blockchain application hit an unrecoverable error:\n\n%v",
					zap.Stringer("error", et))
				// cometbft *may* already recover panics from the application. Investigate.
			default:
				s.log.Error("kwild server panic", zap.Any("error", err))
			}
			*/
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

		s.log.Info("http server started", zap.String("address", s.cfg.AppCfg.HTTPListenAddress))
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
