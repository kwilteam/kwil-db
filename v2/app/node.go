package app

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"kwil/log"
	"kwil/node"
	"kwil/node/types"
	"kwil/version"

	"github.com/libp2p/go-libp2p/core/crypto" // TODO: isolate to config package not main
)

func runNode(ctx context.Context, rootDir string, logLevel log.Level, logFormat log.Format) error {
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

	logger := log.New(log.WithLevel(logLevel), log.WithFormat(logFormat),
		log.WithName("KWILD"), log.WithWriter(logWriter))
	// NOTE: level and name can be set independently for different systems

	logger.Infof("Starting kwild version %v", version.KwilVersion)

	genFile := filepath.Join(rootDir, "genesis.json")

	logger.Infof("Loading the genesis configuration from %s", genFile)

	genConfig, err := node.LoadGenesisConfig(genFile)
	if err != nil {
		return fmt.Errorf("failed to load genesis config: %w", err)
	}

	// assuming static validators
	valSet := make(map[string]types.Validator)
	for _, val := range genConfig.Validators {
		valSet[hex.EncodeToString(val.PubKey)] = val
	}

	cfgFile := filepath.Join(rootDir, "config.json")
	cfg, err := node.LoadNodeConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load node config: %w", err)
	}

	logger.Infof("Loading the node configuration from %s", cfgFile)

	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(cfg.PrivateKey)
	if err != nil {
		return err
	}
	pubKey, err := privKey.GetPublic().Raw()
	if err != nil {
		return err
	}

	logger.Info("Parsing the pubkey", "key", hex.EncodeToString(pubKey))
	// Check if u are the leader
	var nRole types.Role
	if bytes.Equal(pubKey, genConfig.Leader) {
		nRole = types.RoleLeader
		logger.Info("You are the leader")
	} else {
		// check if you are a validator
		if _, ok := valSet[hex.EncodeToString(pubKey)]; ok {
			nRole = types.RoleValidator
			logger.Info("You are a validator")
		} else {
			nRole = types.RoleSentry
			logger.Info("You are a sentry")
		}
	}

	// TODOs:
	//  - node.WithGenesisConfig instead of WithGenesisValidators
	//  - change node.WithPrivateKey to []byte or our own PrivateKey type

	nodeLogger := logger.NewWithLevel(logLevel, "NODE")
	node, err := node.NewNode(rootDir, node.WithPort(cfg.Port), node.WithPrivKey(cfg.PrivateKey),
		node.WithRole(nRole), node.WithPex(cfg.Pex), node.WithGenesisValidators(valSet), node.WithLogger(nodeLogger))
	if err != nil {
		return err
	}

	addr := node.Addr()
	logger.Infof("To connect: %s", addr)

	var bootPeers []string
	if cfg.SeedNode != "" {
		bootPeers = append(bootPeers, cfg.SeedNode)
	}
	if err = node.Start(ctx, bootPeers...); err != nil {
		return err
	}
	// Start is blocking, for now.

	return nil
}
