package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"p2p/log"
	"p2p/node"
	"p2p/node/types"

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
	rootDir string
	// key       string
	// port      uint64
	// connectTo string
	// noPex     bool
	// leader    bool
	// numVals   int
	logLevel  string
	logFormat string
)

func run(ctx context.Context) error {
	// TODO: restore flags, rethink config file (format and/or existence)

	// flag.StringVar(&key, "key", "", "private key bytes (hexadecimal), empty is pseudo-random")
	// flag.Uint64Var(&port, "port", 0, "listen port (0 for random)")
	// flag.StringVar(&connectTo, "connect", "", "peer multiaddr to connect to")
	// flag.BoolVar(&noPex, "no-pd", false, "disable peer discover")
	// flag.BoolVar(&leader, "leader", false, "make this node produce blocks (should only be one in a network)")
	// flag.IntVar(&role, "role", 0, "role of the node (0: leader, 1: validator, 2: sentry)")
	// flag.IntVar(&numVals, "v", 1, "number of validators (all peers are validators!)")
	flag.StringVar(&rootDir, "root", ".testnet", "root directory for the configuration")
	flag.StringVar(&logLevel, "log-level", log.LevelInfo.String(), "log level")
	flag.StringVar(&logFormat, "log-format", string(log.FormatUnstructured), "log format")
	flag.Parse()

	logLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	logger := log.New(log.WithLevel(logLevel), log.WithFormat(log.FormatUnstructured),
		log.WithName("kwild2"))
	// NOTE: level can be set independently for different system

	// dummyce.NumValidatorsFake = numVals

	// rr := rand.Reader
	// if port != 0 { // deterministic key based on port for testing
	// 	// rr = mrand.New(mrand.NewSource(int64(port)))
	// 	var seed [32]byte
	// 	binary.LittleEndian.PutUint64(seed[:], port)
	// 	seed = sha256.Sum256(seed[:])
	// 	log.Printf("seed: %x", seed)
	// 	rr = mrand2.NewChaCha8(seed)
	// 	// var buf bytes.Buffer
	// 	// buf.Write(seed[:])
	// 	// buf.Write(seed[:])
	// 	// rr = &buf
	// }

	// var rawKey []byte
	// if key == "" {
	// 	privKey := node.NewKey(rr)
	// 	rawKey, _ = privKey.Raw()
	// 	log.Printf("priv key: %x", rawKey)
	// } else {
	// 	var err error
	// 	rawKey, err = hex.DecodeString(key)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	fmt.Println("Parsing the configuration")

	valSet := make(map[string]types.Validator)
	genFile := filepath.Join(rootDir, "genesis.json")
	if _, err := os.Stat(genFile); os.IsNotExist(err) {
		return fmt.Errorf("genesis file not found: %s", genFile)
	}

	fmt.Println("Loading the genesis configuration", genFile)

	genConfig, err := node.LoadGenesisConfig(genFile)
	if err != nil {
		return err
	}

	// assuming static validators
	for _, val := range genConfig.Validators {
		valSet[hex.EncodeToString(val.PubKey)] = val
	}

	cfgFile := filepath.Join(rootDir, "config.json")
	cfg, err := node.LoadNodeConfig(cfgFile)
	if err != nil {
		return err
	}

	fmt.Println("Loading the node configuration", cfgFile)

	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(cfg.PrivateKey)
	if err != nil {
		return err
	}
	pubKey, err := privKey.GetPublic().Raw()
	if err != nil {
		return err
	}

	fmt.Println("Parsing the pubkey key", hex.EncodeToString(pubKey))
	// Check if u are the leader
	var nRole types.Role
	if bytes.Equal(pubKey, genConfig.Leader) {
		nRole = types.RoleLeader
		fmt.Println("You are the leader")
	} else {
		// check if you are a validator
		if _, ok := valSet[hex.EncodeToString(pubKey)]; ok {
			nRole = types.RoleValidator
			fmt.Println("You are a validator")
		} else {
			nRole = types.RoleSentry
			fmt.Println("You are a sentry")
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
	logger.Infof("to connect: %s", addr)

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
