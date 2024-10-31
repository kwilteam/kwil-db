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

func runNode(ctx context.Context, rootDir string, cfg *node.Config) error {
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

	genFile := filepath.Join(rootDir, GenesisFileName)

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
	// TODO:  the role assignement should be based on the current valset rather than the genesis config, once we have the persisted valset
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

	ip, port := cfg.PeerConfig.IP, cfg.PeerConfig.Port

	nodeLogger := logger.NewWithLevel(cfg.LogLevel, "NODE")
	// wtf functional options; we're making a config struct soon
	node, err := node.NewNode(rootDir, node.WithIP(ip), node.WithPort(port),
		node.WithPrivKey(cfg.PrivateKey[:]), node.WithLeader(genConfig.Leader[:]),
		node.WithRole(nRole), node.WithGenesisValidators(valSet),
		node.WithPex(cfg.PeerConfig.Pex), node.WithLogger(nodeLogger))
	if err != nil {
		return err
	}

	addr := node.Addr()
	logger.Infof("To connect: %s", addr)

	var bootPeers []string
	if cfg.PeerConfig.BootNode != "" {
		bootPeers = append(bootPeers, cfg.PeerConfig.BootNode)
	}
	if err = node.Start(ctx, bootPeers...); err != nil {
		return err
	}
	// Start is blocking, for now.

	return nil
}
