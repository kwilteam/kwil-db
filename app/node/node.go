package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node"
	"github.com/kwilteam/kwil-db/version"
)

func runNode(ctx context.Context, rootDir string, cfg *config.Config) error {
	// Writing to stdout and a log file.  TODO: config outputs
	rot, err := log.NewRotatorWriter(filepath.Join(rootDir, "kwild.log"), 10_000, 0)
	if err != nil {
		return fmt.Errorf("failed to create log rotator: %w", err)
	}
	defer func() {
		if err := rot.Close(); err != nil {
			fmt.Printf("failed to close log rotator: %v", err)
		}
	}()

	logWriter := io.MultiWriter(os.Stdout, rot) // tee to stdout and log file

	logger := log.New(log.WithLevel(cfg.LogLevel), log.WithFormat(cfg.LogFormat),
		log.WithName("KWILD"), log.WithWriter(logWriter))
	// NOTE: level and name can be set independently for different systems

	logger.Infof("Starting kwild version %v", version.KwilVersion)

	genFile := filepath.Join(rootDir, config.GenesisFileName)

	logger.Infof("Loading the genesis configuration from %s", genFile)

	genConfig, err := config.LoadGenesisConfig(genFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis config: %w", err)
	}

	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(cfg.PrivateKey)
	if err != nil {
		return err
	}
	pubKey := privKey.Public().Bytes()

	logger.Info("Parsing the pubkey", "key", hex.EncodeToString(pubKey))

	nodeCfg := &node.Config{
		RootDir:   rootDir,
		PrivKey:   privKey,
		Logger:    logger.NewWithLevel(cfg.LogLevel, "NODE"),
		Consensus: cfg.Consensus,
		Genesis:   *genConfig,
		P2P:       cfg.P2P,
		PG:        cfg.DB,
	}
	node, err := node.NewNode(nodeCfg)
	if err != nil {
		return err
	}

	addrs := node.Addrs()
	logger.Infof("This node is %s", addrs)

	if err = node.Start(ctx, cfg.P2P.BootNodes...); err != nil {
		return err
	}
	// Start is blocking, for now.

	return nil
}
