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
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node"

	"github.com/spf13/cobra"
)

// accept: #vals, #n-vals and set the first validator as the leader.

func TestnetCmd() *cobra.Command {
	var numVals, numNVals int
	var noPex bool
	var startingPort uint64
	var outDir string

	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Generate configuration for multiple nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenerateTestnetConfigs(outDir, numVals, numNVals, noPex, startingPort)
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

	return cmd
}

func GenerateTestnetConfigs(outDir string, numVals, numNVals int, noPex bool, startingPort uint64) error {
	// ensure that the directory exists
	// expand the directory path
	outDir, err := node.ExpandPath(outDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	var keys []crypto.PrivateKey
	// generate the configuration for the nodes
	for i := range numVals + numNVals {
		// generate Keys, so that the connection strings and the validator set can be generated before the node config files are generated
		var seed [32]byte
		binary.LittleEndian.PutUint64(seed[:], startingPort+uint64(i))
		seed = sha256.Sum256(seed[:])
		rr := rand.NewChaCha8(seed)
		priv := node.NewKey(&deterministicPRNG{ChaCha8: rr})
		keys = append(keys, priv)
	}

	// key 0 is leader
	leaderPub := keys[0].Public()

	genConfig := &config.GenesisConfig{
		ChainID:          "kwil-testnet",
		Leader:           leaderPub.Bytes(), // rethink this so it can be different key types?
		Validators:       make([]*ktypes.Validator, numVals),
		DisabledGasCosts: true,
		JoinExpiry:       14400,
		VoteExpiry:       108000,
		MaxBlockSize:     6 * 1024 * 1024,
		MaxVotesPerTx:    200,
	}

	for i := range numVals {
		genConfig.Validators[i] = &ktypes.Validator{
			PubKey: keys[i].Public().Bytes(),
			Power:  1,
		}
	}

	// generate the configuration for the nodes
	for i := range numVals + numNVals {
		err = GenerateNodeRoot(&NodeGenConfig{
			PortOffset: i,
			IP:         "127.0.0.1",
			NoPEX:      noPex,
			RootDir:    filepath.Join(outDir, fmt.Sprintf("node%d", i)),
			NodeKey:    keys[i],
			Genesis:    genConfig,
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
	NodeKey    crypto.PrivateKey
	Genesis    *config.GenesisConfig

	// TODO: gasEnabled, private p2p, auth RPC, join expiry, allocs, etc.
}

func GenerateNodeRoot(ncfg *NodeGenConfig) error {
	if err := os.MkdirAll(ncfg.RootDir, 0755); err != nil {
		return err
	}

	cfg := custom.DefaultConfig() // not config.DefaultConfig(), so custom command config is used

	// P2P
	cfg.P2P.Port = uint64(ncfg.PortOffset + 6600)
	cfg.P2P.IP = "127.0.0.1"
	if ncfg.IP != "" {
		cfg.P2P.IP = ncfg.IP
	}
	cfg.P2P.Pex = !ncfg.NoPEX

	leaderPub, err := crypto.UnmarshalPublicKey(ncfg.Genesis.Leader, crypto.KeyTypeSecp256k1)
	if err != nil {
		return err
	}

	if !ncfg.NodeKey.Public().Equals(leaderPub) {
		// make everyone connect to leader
		cfg.P2P.BootNodes = []string{node.FormatPeerString(
			leaderPub.Bytes(), leaderPub.Type(), cfg.P2P.IP, 6600)}
	}

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

	if err := cfg.SaveAs(filepath.Join(ncfg.RootDir, config.ConfigFileName)); err != nil {
		return err
	}

	// save the genesis configuration to the root directory
	genFile := filepath.Join(ncfg.RootDir, config.GenesisFileName)
	if err := ncfg.Genesis.SaveAs(genFile); err != nil {
		return err
	}

	return key.SaveNodeKey(filepath.Join(ncfg.RootDir, "nodekey.json"), ncfg.NodeKey)
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
