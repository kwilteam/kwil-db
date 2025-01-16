package setup

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/key"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node"

	"github.com/spf13/cobra"
)

// accept: #vals, #n-vals and set the first validator as the leader.

func TestnetCmd() *cobra.Command {
	var numVals, numNVals int
	var noPex, uniquePorts bool
	var startingPort uint64
	var outDir, dbOwner string

	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Generate configuration for multiple nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenerateTestnetConfigs(&TestnetConfig{
				RootDir:      outDir,
				NumVals:      numVals,
				NumNVals:     numNVals,
				NoPex:        noPex,
				StartingPort: startingPort,
				Owner:        dbOwner,
			}, &ConfigOpts{
				UniquePorts: uniquePorts,
				DnsHost:     false,
			})
		},
	}

	// root dir does not apply to this command. NOTE: probably this means we
	// need a top level "node" command with root and all the other config binds.
	// On the other hand, `setup reset` applies to a specific node instance...
	cmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		cmd.InheritedFlags().MarkHidden(bind.RootFlagName)
		cmd.Parent().HelpFunc()(c, s)
	})

	cmd.Flags().IntVarP(&numVals, "vals", "v", 3, "number of validators (includes the one leader)")
	cmd.Flags().IntVarP(&numNVals, "non-vals", "n", 0, "number of non-validators (default 0)")
	cmd.Flags().BoolVar(&noPex, "no-pex", false, "disable peer exchange")
	cmd.Flags().Uint64VarP(&startingPort, "port", "p", 6600, "starting P2P port for the nodes")
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", ".testnet", "output directory for generated node root directories")
	cmd.Flags().BoolVarP(&uniquePorts, "unique-ports", "u", false, "use unique ports for each node")
	cmd.Flags().StringVar(&dbOwner, "db-owner", "", "owner of the database")
	return cmd
}

type TestnetConfig struct {
	RootDir       string
	ChainID       string
	NumVals       int
	NumNVals      int
	NoPex         bool
	StartingPort  uint64
	StartingIP    string
	DnsNamePrefix string // optional and only used if DnsHost is true (default: node)

	Owner string
}

type ConfigOpts struct {
	// UniquePorts is a flag to generate unique listening addresses
	// (JSON-RPC, HTTP, Admin, P2P, node RPC) for each node.
	// This is useful for testing multiple nodes on the same machine.
	// If it is used for generating a single config, it has no effect.
	UniquePorts bool

	// DnsHost is a flag to use DNS hostname as host in the config
	// instead of ip. It will be used together with DnsNamePrefix to generate
	// hostnames.
	// This is useful for testing nodes inside docker containers.
	DnsHost bool
}

// TODO: once changes to the tests are complete, this may not be needed
func GenerateTestnetConfigs(cfg *TestnetConfig, opts *ConfigOpts) error {
	// ensure that the directory exists
	// expand the directory path
	outDir, err := node.ExpandPath(cfg.RootDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	var keys []*crypto.Secp256k1PrivateKey
	// generate the configuration for the nodes
	for i := range cfg.NumVals + cfg.NumNVals {
		// generate Keys, so that the connection strings and the validator set can be generated before the node config files are generated
		var seed [32]byte
		binary.LittleEndian.PutUint64(seed[:], cfg.StartingPort+uint64(i))
		seed = sha256.Sum256(seed[:])
		rr := rand.NewChaCha8(seed)
		priv := node.NewKey(&deterministicPRNG{ChaCha8: rr})
		keys = append(keys, priv)
	}

	// key 0 is leader
	leaderPub := keys[0].Public()

	var bootNodes []string
	for i := range cfg.NumVals {
		pubKey := keys[i].Public()

		hostname := cfg.StartingIP
		if cfg.StartingIP == "" {
			hostname = "127.0.0.1"
		}

		if opts.DnsHost {
			hostname = fmt.Sprintf("%s%d", cfg.DnsNamePrefix, i)
		}

		port := 6600
		if opts.UniquePorts {
			port = 6600 + i
		}

		bootNodes = append(bootNodes, node.FormatPeerString(pubKey.Bytes(), pubKey.Type(), hostname, port))
	}

	chainID := cfg.ChainID
	if chainID == "" {
		chainID = "kwil-testnet"
	}

	genConfig := config.DefaultGenesisConfig()
	genConfig.ChainID = chainID
	genConfig.Leader = types.PublicKey{PublicKey: leaderPub}
	genConfig.Validators = make([]*ktypes.Validator, cfg.NumVals)
	genConfig.DBOwner = cfg.Owner
	if genConfig.DBOwner == "" {
		signer := auth.GetUserSigner(keys[0])
		ident, err := auth.GetIdentifierFromSigner(signer)
		if err != nil {
			return fmt.Errorf("failed to get identifier from user signer for dbOwner: %w", err)
		}
		genConfig.DBOwner = ident
	}

	for i := range cfg.NumVals {
		genConfig.Validators[i] = &ktypes.Validator{
			AccountID: types.AccountID{
				Identifier: keys[i].Public().Bytes(),
				KeyType:    keys[i].Type(),
			},
			Power: 1,
		}
	}

	// generate the configuration for the nodes
	portOffset := 0
	for i := range cfg.NumVals + cfg.NumNVals {
		if opts.UniquePorts {
			portOffset = i
		}
		err = GenerateNodeRoot(&NodeGenConfig{
			PortOffset: portOffset,
			IP:         cfg.StartingIP,
			NoPEX:      cfg.NoPex,
			RootDir:    filepath.Join(outDir, fmt.Sprintf("node%d", i)),
			NodeKey:    keys[i],
			Genesis:    genConfig,
			BootNodes:  bootNodes,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type NodeGenConfig struct {
	RootDir    string
	PortOffset int
	DBPort     uint16 // leave zero for default plus any offset
	IP         string
	NoPEX      bool
	NodeKey    *crypto.Secp256k1PrivateKey
	Genesis    *config.GenesisConfig

	// TODO: gasEnabled, private p2p, auth RPC, join expiry, allocs, etc.
	BootNodes []string
}

func GenerateNodeRoot(ncfg *NodeGenConfig) error {
	cfg := custom.DefaultConfig() // not config.DefaultConfig(), so custom command config is used

	// P2P
	port := uint64(ncfg.PortOffset + 6600)
	host := "127.0.0.1"
	if ncfg.IP != "" {
		host = ncfg.IP
	}
	cfg.P2P.ListenAddress = net.JoinHostPort(host, strconv.FormatUint(port, 10))
	cfg.P2P.Pex = !ncfg.NoPEX

	cfg.P2P.BootNodes = ncfg.BootNodes

	// DB
	dbPort := ncfg.DBPort
	if dbPort == 0 {
		dbPort = uint16(5432 + ncfg.PortOffset)
	}
	cfg.DB.Port = strconv.FormatUint(uint64(dbPort), 10)

	// RPC
	cfg.RPC.ListenAddress = net.JoinHostPort("0.0.0.0", strconv.FormatUint(uint64(8484+ncfg.PortOffset), 10))

	// Admin RPC
	cfg.Admin.ListenAddress = net.JoinHostPort("127.0.0.1", strconv.FormatUint(uint64(8584+ncfg.PortOffset), 10))

	return GenerateNodeDir(ncfg.RootDir, ncfg.Genesis, cfg, ncfg.NodeKey)
}

type deterministicPRNG struct {
	readBuf [8]byte
	readLen int // 0 <= readLen <= 8
	*rand.ChaCha8
}

// Read is a bad replacement for the actual Read method added in Go 1.23
func (dr *deterministicPRNG) Read(p []byte) (n int, err error) {
	// fill p by calling Uint64 in a loop until we have enough bytes
	if dr.readLen > 0 {
		n = copy(p, dr.readBuf[len(dr.readBuf)-dr.readLen:])
		dr.readLen -= n
		p = p[n:]
	}
	for len(p) >= 8 {
		binary.LittleEndian.PutUint64(p, dr.ChaCha8.Uint64())
		p = p[8:]
		n += 8
	}
	if len(p) > 0 {
		binary.LittleEndian.PutUint64(dr.readBuf[:], dr.Uint64())
		n += copy(p, dr.readBuf[:])
		dr.readLen = 8 - len(p)
	}
	return n, nil
}

type TestnetNodeConfig struct {
	// Config is the node configuration.
	Config *config.Config
	// DirName is the directory name of the node.
	// If the testnetDir is "testnet" and the DirName is "node0",
	// the full path of the node is "testnet/node0".
	DirName string
	// PrivateKey is the private key of the node.
	PrivateKey *crypto.Secp256k1PrivateKey
}

// GenerateTestnetDir generates a testnet configuration for multiple nodes.
// It is a minimal function that takes full configurations.
// Most users should use GenerateTestnetConfigs instead.
func GenerateTestnetDir(testnetDir string, genesis *config.GenesisConfig, nodes []*TestnetNodeConfig) error {
	if err := os.MkdirAll(testnetDir, 0755); err != nil {
		return err
	}

	for _, node := range nodes {
		if err := GenerateNodeDir(filepath.Join(testnetDir, node.DirName), genesis, node.Config, node.PrivateKey); err != nil {
			return err
		}
	}

	return nil
}

// GenerateNodeDir generates a node configuration directory.
// It is a minimal function that takes a full configuration.
// Most users should use GenerateNodeRoot instead.
func GenerateNodeDir(rootDir string, genesis *config.GenesisConfig, node *config.Config, privateKey *crypto.Secp256k1PrivateKey) error {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return err
	}

	if err := node.SaveAs(config.ConfigFilePath(rootDir)); err != nil {
		return err
	}

	if err := genesis.SaveAs(config.GenesisFilePath(rootDir)); err != nil {
		return err
	}

	return key.SaveNodeKey(config.NodeKeyFilePath(rootDir), privateKey)
}
