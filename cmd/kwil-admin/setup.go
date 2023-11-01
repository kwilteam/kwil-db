package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"

	"github.com/alexflint/go-arg"
)

// kwil-admin setup init -o DIR ...
// kwil-admin setup testnet -v X -n Y ...
// kwil-admin setup genesis-hash DBDIR --genesis=<../genesis.json>
// kwil-admin setup reset
// kwil-admin setup reset-state

type SetupCmd struct {
	Init        *SetupInitCmd        `arg:"subcommand:init" help:"Initialize the files required for a single Kwil node."`
	Testnet     *SetupTestnetCmd     `arg:"subcommand:testnet" help:"Initialize the files required for a Kwil test network."`
	GenesisHash *SetupGenesisHashCmd `arg:"subcommand:genesis-hash" help:"Compute genesis hash from SQLite files, and optionally update genesis.json"`
	Reset       *SetupResetCmd       `arg:"subcommand:reset" help:"Delete all Kwil data folders, including datasets and blockchain data"`
	ResetState  *SetupResetStateCmd  `arg:"subcommand:reset-state" help:"Delete blockchain state files"`
}

func (cc *SetupCmd) run(ctx context.Context) error {
	switch {
	case cc.Init != nil:
		genCfg := &nodecfg.NodeGenerateConfig{
			ChainID:         cc.Init.ChainID,
			OutputDir:       cc.Init.OutputDir,
			JoinExpiry:      cc.Init.JoinExpiry,
			WithoutGasCosts: true, // gas disabled by setup init
			WithoutNonces:   cc.Init.WithoutNonces,
		}
		return nodecfg.GenerateNodeConfig(genCfg)
	case cc.Testnet != nil:
		cc.Testnet.WithoutGasCosts = true // gas disabled by setup testnet
		genCfg := nodecfg.TestnetGenerateConfig(*cc.Testnet)
		genCfg.PopulatePersistentPeers = true // temporary workaround without changing nodecfg
		return nodecfg.GenerateTestnetConfig(&genCfg)
	case cc.GenesisHash != nil:
		dbDir, genesisFile := cc.GenesisHash.DBDir, cc.GenesisHash.GenesisFile
		appHash, err := config.PatchGenesisAppHash(dbDir, genesisFile)
		if err != nil {
			fmt.Printf("App hash written: %x", appHash)
			return nil
		}
		return err
	case cc.Reset != nil:
		return cc.Reset.run(ctx)
	case cc.ResetState != nil:
		return cc.ResetState.run(ctx)
	default:
		return arg.ErrHelp
	}
}

type SetupInitCmd struct {
	ChainID       string `arg:"--chain-id" help:"override the chain ID"`
	OutputDir     string `arg:"-o,--output-dir" default:".testnet" help:"parent directory for all of generated node folders" placeholder:"DIR"`
	JoinExpiry    int64  `arg:"--join-expiry" default:"86400" help:"number of blocks before a join request expires"`
	WithoutNonces bool   `arg:"--without-nonces" help:"disable nonces"`

	// WithoutGasCosts is not an available flag since Kwil users have no way to
	// get funded with the external chain syncer gone.
	// WithoutGasCosts bool   `arg:"--without-gas-costs" default:"true" help:"disable gas costs"`
}

// SetupTestnetCmd exactly matches nodecfg.TestnetGenerateConfig in field name,
// type, and layout so that it may be converted directly.
type SetupTestnetCmd struct {
	ChainID                 string   `arg:"--chain-id" help:"override the chain ID"`
	NValidators             int      `arg:"-v,--validators" default:"4" help:"number of validators" placeholder:"V"`
	NNonValidators          int      `arg:"-n,--non-validators" default:"4" help:"number of non-validators" placeholder:"N"`
	ConfigFile              string   `arg:"--config" help:"template config file to use, default is none" placeholder:"FILE"`
	OutputDir               string   `arg:"-o,--output-dir" default:".testnet" help:"parent directory for all of generated node folders" placeholder:"DIR"`
	NodeDirPrefix           string   `arg:"--node-dir-prefix" default:"node" help:"prefix for the node directories (node results in node0, node1, ...)" placeholder:"PRE"`
	PopulatePersistentPeers bool     `arg:"-"` // `arg:"--populate-persistent-peers" help:"update config of each node with the list of persistent peers build using either hostname-prefix or starting-ip-address"`
	HostnamePrefix          string   `arg:"--hostname-prefix" help:"prefix for node host names e.g. node results in node0, node1, etc." placeholder:"PRE"`
	HostnameSuffix          string   `arg:"--hostname-suffix" help:"suffix for node host names e.g. .example.com results in node0.example.com, node1.example.com, etc." placeholder:"SUF"`
	StartingIPAddress       string   `arg:"--starting-ip" help:"starting IP address of the first network node" placeholder:"IP"`
	Hostnames               []string `arg:"--hostnames" help:"override all hostnames of the nodes (list of hostnames must be the same length as the number of nodes)" placeholder:"HOST"`
	P2pPort                 int      `arg:"-p,--p2p-port" help:"P2P port" default:"26656" placeholder:"PORT"`
	JoinExpiry              int64    `arg:"--join-expiry" default:"86400" help:"number of blocks before a join request expires"`
	WithoutGasCosts         bool     `arg:"-"` // we force true since kwild doesn't work with gas for this release.
	WithoutNonces           bool     `arg:"--without-nonces" help:"disable nonces"`
}

// TODO: customize the parser to recognize a detailer subcommand and print
// extended details after auto-generated usage. Presently this is not shown by
// WriteHelpForSubcommand.
func (*SetupTestnetCmd) Details() string {
	return `The testnet command creates "v + n" node root directories and populates
each with necessary files to start the new network.

The genesis file includes list of v validators under the validators section.

NOTE: strict routability for addresses is turned off in the config file so that
the test network of nodes can run on a LAN.

Optionally, it will fill in the persistent_peers list in the config file using
either hostnames or IPs.

Examples:

	# Generate a network with 4 validators and 4 non-validators with the IPs
	# 192.168.10.{2,...,9}
	kwil-admin setup testnet -v 4 -o ./output --starting-ip 192.168.10.2

	# Same as above but only 2 additional (non-validator) nodes
	kwil-admin setup testnet -v 4 -n 2 --o ./output --starting-ip 192.168.10.2

	# Manually specify hostnames for the nodes
	kwil-admin setup testnet -v 4 -o ./output --hostnames 192.168.10.2 192.168.10.3 ...
`
}

type SetupGenesisHashCmd struct {
	DBDir       string `arg:"positional" help:"directory containing all of kwild's .sqlite files to be hashed"`
	GenesisFile string `arg:"-g,--genesis" help:"optional path to the genesis file to patch with the computed app hash"`
}

func (*SetupGenesisHashCmd) Details() string {
	return `Generate the genesis hash of the sqlite DBs and if a genesis file is provided,
update the app_hash in the genesis file. If genesis file is not provided, only
print the the genesis hash to stdout.`
}

type resetDirCmd struct {
	RootDir string `arg:"--root_dir,-r" help:"Kwil server root directory" placeholder:"DIR"`
	Force   bool   `arg:"-f,--force" help:"remove the default home without explicit specification"`
	// KeepAddrBook bool   `arg:"-k,--keep-addrbook" help:"keep the address book intact"`
}

// rootDir gets the root directory from the RootDir field first, and if it is
// not set AND Force is true, then it will use the default root of ~/.kwild.
func (rdc *resetDirCmd) rootDir() (string, error) {
	rootDir := rdc.RootDir
	if rootDir == "" {
		if !rdc.Force {
			return "", errors.New("not removing default home directory without --force or --root_dir")
		}
		rootDir = defaultKwildRoot()
	}
	return rootDir, nil
}

type SetupResetStateCmd struct {
	resetDirCmd
}

func (src *SetupResetStateCmd) run(_ context.Context) error {
	rootDir, err := src.rootDir()
	if err != nil {
		return err
	}
	return config.ResetChainState(rootDir)
}

type SetupResetCmd struct {
	resetDirCmd

	SQLitePath  string `arg:"--sqlpath" placeholder:"DIR" help:"Path to the SQLite files"`
	SnapshotDir string `arg:"--snappath" placeholder:"DIR" help:"Path to the snapshots"`
}

func (src *SetupResetCmd) run(_ context.Context) error {
	rootDir, err := src.rootDir()
	if err != nil {
		return err
	}
	sqlitePath, snapshotDir := src.SQLitePath, src.SnapshotDir
	if sqlitePath == "" {
		sqlitePath = config.DefaultSQLitePath
	}
	if snapshotDir == "" {
		snapshotDir = config.DefaultSnapshotsDir
	}
	return config.ResetAll(rootDir, sqlitePath, snapshotDir)
}
