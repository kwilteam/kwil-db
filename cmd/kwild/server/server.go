// Package server defines the main Kwil server, which includes the blockchain
// node and the gRPC services that interface with the Kwil dataset engine.
package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	kconfig "github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/chain"
	config "github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/listeners"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/version"
)

// Server controls the gRPC server and http gateway.
type Server struct {
	jsonRPCServer      *rpcserver.Server
	jsonRPCAdminServer *rpcserver.Server
	cometBftNode       *cometbft.CometBftNode
	listenerManager    *listeners.ListenerManager
	closers            *closeFuncs
	log                log.Logger

	dbCtx interface {
		Done() <-chan struct{}
		Err() error
	}

	cfg *config.KwildConfig

	cancelCtxFunc context.CancelFunc
}

const (
	// Top-level directory structure for the Server's systems
	abciDirName    = kconfig.ABCIDirName
	signingDirName = kconfig.SigningDirName
)

// New builds the kwild server.
func New(ctx context.Context, cfg *config.KwildConfig, genesisCfg *chain.GenesisConfig,
	nodeKey *crypto.Ed25519PrivateKey, autogen bool) (svr *Server, err error) {
	logCfg, err := cfg.LogConfig()
	if err != nil {
		return nil, err
	}

	logger, err := log.NewChecked(*logCfg)
	if err != nil {
		return nil, fmt.Errorf("invalid logger config: %w", err)
	}
	logger = *logger.Named("kwild")

	closers := &closeFuncs{
		closers: []func() error{}, // logger.Close is not in here; do it in a defer in Start
		logger:  logger,
	}

	defer func() {
		if r := recover(); r != nil {
			svr = nil
			if pe, ok := r.(panicErr); ok && errors.Is(pe.err, context.Canceled) {
				logger.Warnf("Shutdown signaled: %v", pe.msg)
				err = pe // interrupt request (shutdown) during bringup, not a crash
			} else {
				stack := make([]byte, 8192)
				length := runtime.Stack(stack, false)
				err = fmt.Errorf("panic while building kwild: %v\n\nstack:\n\n%v", r, string(stack[:length]))
			}
			closers.closeAll()
		}
	}()

	if cfg.AppConfig.TLSKeyFile == "" || cfg.AppConfig.TLSCertFile == "" {
		return nil, errors.New("unspecified TLS key and/or certificate")
	}

	if err = os.MkdirAll(cfg.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory %q: %w", cfg.RootDir, err)
	}

	logger.Debug("loading TLS key pair for gRPC servers", log.String("key_file", "d.cfg.TLSKeyFile"),
		log.String("cert_file", "d.cfg.TLSCertFile")) // wtf why can't we log yet?
	keyPair, err := loadTLSCertificate(cfg.AppConfig.TLSKeyFile, cfg.AppConfig.TLSCertFile, cfg.AppConfig.Hostname)
	if err != nil {
		return nil, err
	}

	dbLogger := increaseLogLevel("pg", &logger, cfg.Logging.DBLevel)
	pg.UseLogger(*dbLogger)

	host, port, user, pass := cfg.AppConfig.DBHost, cfg.AppConfig.DBPort, cfg.AppConfig.DBUser, cfg.AppConfig.DBPass

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
		if err := recover(); err != nil {
			s.log.Error("kwild server panic", zap.Any("error", err))
		}

		s.log.Info("Closing server resources...")
		err := s.closers.closeAll()
		if err != nil {
			s.log.Error("failed to close resource:", zap.Error(err))
		}
		s.log.Info("Server is now shut down.")
		s.log.Close()
	}()

	s.log.Infof("Starting server (kwild version %v)...", version.KwilVersion)

	cancelCtx, done := context.WithCancel(ctx)
	s.cancelCtxFunc = done

	group, groupCtx := errgroup.WithContext(cancelCtx)

	group.Go(func() error {
		// If the DB dies unexpectedly, stop the entire error group.
		select {
		case <-s.dbCtx.Done(): // DB died
			return s.dbCtx.Err() // shutdown the server
		case <-groupCtx.Done(): // something else died or was shut down
			return nil
		}
	})

	group.Go(func() error {
		s.log.Info("starting user json-rpc server", zap.String("address", s.cfg.AppConfig.JSONRPCListenAddress))
		return s.jsonRPCServer.Serve(groupCtx)
	})

	group.Go(func() error {
		s.log.Info("starting admin json-rpc server", zap.String("address", s.cfg.AppConfig.AdminListenAddress))
		return s.jsonRPCAdminServer.Serve(groupCtx)
	})

	// Start listener manager only after node caught up
	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop listeners")
			s.listenerManager.Stop()
		}()
		return s.listenerManager.Start()
	})
	s.log.Info("listener manager started")

	group.Go(func() error {
		// The CometBFT services do not block on Start().
		if err := s.cometBftNode.Start(); err != nil {

			return err
		}
		// If you create DB errors from start, note that this needs DB writes
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

		s.log.Error("server error", zap.Error(err))
		s.cancelCtxFunc()
		return err
	}

	return nil
}

// Stop begins shutting down the Server. However, the caller of Start will
// normally cancel the provided context and wait for it to return.
func (s *Server) Stop() {
	s.log.Warn("stop kwild services")
	s.cancelCtxFunc()
}
