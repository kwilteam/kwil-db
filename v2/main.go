package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"p2p/log"
	"p2p/node"
	"p2p/node/types"

	// TODO: isolate to config package not main
	"github.com/libp2p/go-libp2p/core/crypto" // TODO: isolate to config package not main
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("shutdown signal received")
		cancel()
	}()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

var (
	rootDir   string
	logLevel  string
	logFormat string
)

func run(ctx context.Context) error {
	// TODO: restore flags, rethink config file (format and/or existence)
	flag.StringVar(&rootDir, "root", ".testnet", "root directory for the configuration")
	flag.StringVar(&logLevel, "log-level", log.LevelInfo.String(), "log level")
	flag.StringVar(&logFormat, "log-format", string(log.FormatUnstructured), "log format")
	flag.Parse()

	logLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	logFormat, err := log.ParseFormat(logFormat)
	if err != nil {
		return fmt.Errorf("invalid log format: %w", err)
	}

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
