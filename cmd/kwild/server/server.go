// Package server defines the main Kwil server, which includes the blockchain
// node and the gRPC services that interface with the Kwil dataset engine.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/oracles"
	gateway "github.com/kwilteam/kwil-db/internal/services/grpc_gateway"
	grpc "github.com/kwilteam/kwil-db/internal/services/grpc_server"
	"github.com/kwilteam/kwil-db/internal/sql/pg"

	// internalize
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Server controls the gRPC server and http gateway.
type Server struct {
	grpcServer     *grpc.Server
	gateway        *gateway.GatewayServer
	adminTPCServer *grpc.Server
	cometBftNode   *cometbft.CometBftNode
	oracleMgr      *oracles.OracleMgr
	closers        *closeFuncs
	log            log.Logger

	cfg *config.KwildConfig

	cancelCtxFunc context.CancelFunc
}

const (
	// Top-level directory structure for the Server's systems
	abciDirName        = config.ABCIDirName
	applicationDirName = config.ApplicationDirName
	rcvdSnapsDirName   = config.ReceivedSnapsDirName
	signingDirName     = config.SigningDirName

	// Note that the sqlLite file path is user-configurable
	// e.g. "data/kwil.db"
)

// New builds the kwild server.
func New(ctx context.Context, cfg *config.KwildConfig, genesisCfg *config.GenesisConfig, nodeKey *crypto.Ed25519PrivateKey, autogen bool) (svr *Server, err error) {
	closers := &closeFuncs{
		closers: make([]func() error, 0),
	}

	defer func() {
		if r := recover(); r != nil {
			svr = nil
			stack := make([]byte, 8192)
			length := runtime.Stack(stack, false)
			err = fmt.Errorf("panic while building kwild: %v\n\nstack:\n\n%v", r, string(stack[:length]))
			closers.closeAll()
		}
	}()

	logger, err := log.NewChecked(*cfg.LogConfig())
	if err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}
	logger = *logger.Named("kwild")

	if cfg.AppCfg.TLSKeyFile == "" || cfg.AppCfg.TLSCertFile == "" {
		return nil, errors.New("unspecified TLS key and/or certificate")
	}

	if err = os.MkdirAll(cfg.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory %q: %w", cfg.RootDir, err)
	}

	logger.Debug("loading TLS key pair for gRPC servers", zap.String("key_file", "d.cfg.TLSKeyFile"),
		zap.String("cert_file", "d.cfg.TLSCertFile")) // wtf why can't we log yet?
	keyPair, err := loadTLSCertificate(cfg.AppCfg.TLSKeyFile, cfg.AppCfg.TLSCertFile, cfg.AppCfg.Hostname)
	if err != nil {
		return nil, err
	}

	pg.UseLogger(*logger.Named("pg"))

	host, port, user, pass := cfg.AppCfg.DBHost, cfg.AppCfg.DBPort, cfg.AppCfg.DBUser, cfg.AppCfg.DBPass

	deps := &coreDependencies{
		ctx:        ctx,
		autogen:    autogen,
		cfg:        cfg,
		genesisCfg: genesisCfg,
		privKey:    ed25519.PrivKey(nodeKey.Bytes()),
		log:        logger,
		dbOpener:   newDBOpener(host, port, user, pass), // could make cfg.AppCfg.DBName baked into it this one too
		poolOpener: newPoolBOpener(host, port, user, pass),
		keypair:    keyPair,
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
			s.log.Info("stop admin server")
			s.adminTPCServer.Stop()
		}()

		return s.adminTPCServer.Start()
	})
	s.log.Info("grpc server started", zap.String("address", s.cfg.AppCfg.AdminListenAddress))

	// Start oracle manager only after node caught up
	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop oracle manager")
			s.oracleMgr.Stop()
		}()

		s.oracleMgr.Start()
		return nil
	})
	s.log.Info("oracle manager started")

	group.Go(func() error {
		// The CometBFT services do not block on Start().
		if err := s.cometBftNode.Start(); err != nil {

			return err
		}
		// If you create DB errors from start, note that this is neds db writes
		// in InitChain before transactional block processing begins! Further,
		// it will immediately start replaying blocks if ABCI app indicates it
		// is behind, causing FinalizeBlock+Commit calls right away.
		s.log.Info("comet node is now started")

		<-groupCtx.Done()
		s.log.Info("stop comet server")
		if err := s.cometBftNode.Stop(); err != nil {
			return fmt.Errorf("failed to stop comet server: %w", err)
		}
		s.log.Info("comet server is stopped")
		return nil
	})

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
