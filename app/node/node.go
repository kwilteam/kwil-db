package node

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/kwilteam/kwil-db/app/key"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/listeners"
	rpcserver "github.com/kwilteam/kwil-db/node/services/jsonrpc"
	"github.com/kwilteam/kwil-db/version"

	"golang.org/x/sync/errgroup"
)

type server struct {
	cfg *config.Config // KwildConfig
	log log.Logger

	cancelCtxFunc context.CancelFunc
	closers       *closeFuncs
	dbCtx         interface {
		Done() <-chan struct{}
		Err() error
	}

	// subsystems
	node               *node.Node
	ce                 *consensus.ConsensusEngine
	listeners          *listeners.ListenerManager
	jsonRPCServer      *rpcserver.Server
	jsonRPCAdminServer *rpcserver.Server
}

func runNode(ctx context.Context, rootDir string, cfg *config.Config) (err error) {
	var logWriters []io.Writer
	if idx := slices.Index(cfg.LogOutput, "stdout"); idx != -1 {
		logWriters = append(logWriters, os.Stdout)
		cfg.LogOutput = slices.Delete(cfg.LogOutput, idx, idx+1)
	}
	if idx := slices.Index(cfg.LogOutput, "stderr"); idx != -1 {
		logWriters = append(logWriters, os.Stderr)
		cfg.LogOutput = slices.Delete(cfg.LogOutput, idx, idx+1)
	}

	for _, logFile := range cfg.LogOutput {
		rootedLogFile := rootedPath(logFile, rootDir)
		rot, err := log.NewRotatorWriter(rootedLogFile, 10_000, 0)
		if err != nil {
			return fmt.Errorf("failed to create log rotator: %w", err)
		}
		defer func() {
			if err := rot.Close(); err != nil {
				fmt.Printf("failed to close log rotator: %v", err)
			}
		}()
		logWriters = append(logWriters, rot)
	}

	logger := log.DiscardLogger
	if len(logWriters) > 0 {
		logWriter := io.MultiWriter(logWriters...)

		logger = log.New(log.WithLevel(cfg.LogLevel), log.WithFormat(cfg.LogFormat),
			log.WithName("KWILD"), log.WithWriter(logWriter))
		// NOTE: level and name can be set independently for different systems
	}

	logger.Infof("Starting kwild version %v", version.KwilVersion)

	genFile := rootedPath(config.GenesisFileName, rootDir)

	logger.Infof("Loading the genesis configuration from %s", genFile)

	genConfig, err := config.LoadGenesisConfig(genFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis config: %w", err)
	}

	keyFilePath := filepath.Join(rootDir, "nodekey.json")
	privKey, err := key.LoadNodeKey(keyFilePath)
	if err != nil {
		return fmt.Errorf("failed to load node key: %w", err)
	}
	pubKey := privKey.Public().Bytes()
	logger.Infoln("Node public key:", hex.EncodeToString(pubKey))

	var tlsKeyPair *tls.Certificate
	logger.Info("loading TLS key pair for the admin server", "key_file", cfg.Admin.TLSKeyFile,
		"cert_file", cfg.Admin.TLSKeyFile)
	if cfg.Admin.TLSKeyFile != "" || cfg.Admin.TLSCertFile != "" {
		customHostname := "" // cfg TODO
		keyFile := rootedPath(cfg.Admin.TLSKeyFile, rootDir)
		certFile := rootedPath(cfg.Admin.TLSCertFile, rootDir)
		tlsKeyPair, err = loadTLSCertificate(keyFile, certFile, customHostname)
		if err != nil {
			return err
		}
	}

	host, port, user, pass := cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Pass

	// Prepare for migration if node is setup to use zero-downtime migrations
	// This includes fetching the genesis state from the nodes in the old chain
	if cfg.Migrations.Enable {
		cfg, genConfig, err = prepareForMigration(ctx, cfg, genConfig, rootDir, logger)
		if err != nil {
			return fmt.Errorf("failed to prepare for migration: %w", err)
		}
	}

	if err := genConfig.SanityChecks(); err != nil {
		return fmt.Errorf("genesis configuration failed sanity checks: %w", err)
	}

	d := &coreDependencies{
		ctx:        ctx,
		rootDir:    rootDir,
		adminKey:   tlsKeyPair,
		cfg:        cfg,
		genesisCfg: genConfig,
		privKey:    privKey,
		logger:     logger,
		dbOpener:   newDBOpener(host, port, user, pass),
		poolOpener: newPoolBOpener(host, port, user, pass),
	}

	// Catch any panic from buildServer. We use a panic based build failure
	// system, which is detectable with panicErr. Spewing out a stack is sloppy
	// and a bad look; just return a meaningful error message. If the message is
	// ambiguous regarding its source, the errors needs more context.
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(panicErr); ok { // error on bringup, not a bug
				if errors.Is(pe.err, context.Canceled) {
					logger.Warnf("Shutdown signaled: %v", pe.msg)
					err = context.Canceled
				} else {
					err = pe
				}
			} else { // actual panic (bug)
				stack := make([]byte, 8192)
				length := runtime.Stack(stack, false)
				err = fmt.Errorf("panic while building kwild: %v\n\nstack:\n\n%v", r, string(stack[:length]))
			}
			d.closers.closeAll()
		}
	}()

	server := buildServer(ctx, d)

	return server.Start(ctx)
}

func (s *server) Start(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			s.log.Error("Panic in server", "error", err)
		}

		s.log.Info("Closing server resources...")
		err := s.closers.closeAll()
		if err != nil {
			s.log.Error("failed to close resource:", "error", err)
		}
		s.log.Info("Server is now shut down.")
	}()

	s.log.Info("Starting the server")

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

	// start rpc services
	group.Go(func() error {
		s.log.Info("starting user json-rpc server", "listen", s.cfg.RPC.ListenAddress)
		return s.jsonRPCServer.Serve(groupCtx)
	})

	if s.cfg.Admin.Enable {
		group.Go(func() error {
			s.log.Info("starting admin json-rpc server", "listen", s.cfg.Admin.ListenAddress)
			return s.jsonRPCAdminServer.Serve(groupCtx)
		})
	}

	// start node (p2p)
	group.Go(func() error {
		if err := s.node.Start(groupCtx, s.cfg.P2P.BootNodes...); err != nil {
			s.log.Error("failed to start node", "error", err)
			s.cancelCtxFunc() // Ensure all services are stopped
			return err
		}
		return nil
	})

	// Start listener manager
	group.Go(func() error {
		go func() {
			<-groupCtx.Done()
			s.log.Info("stop listeners")
			s.listeners.Stop()
		}()
		return s.listeners.Start()
	})
	s.log.Info("listener manager started")

	// TODO: node is starting the consensus engine for ease of testing
	// Start the consensus engine

	err := group.Wait()

	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.log.Info("server context is canceled")
			return nil
		}

		s.log.Error("server error", "error", err)
		s.cancelCtxFunc()
		return err
	}

	return nil
}

// rootedPath returns an absolute path for the given path, relative to the root
// directory if it was a relative path.
func rootedPath(path, rootDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(rootDir, path)
}
