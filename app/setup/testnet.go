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
			return generateNodeConfig(outDir, numVals, numNVals, noPex, startingPort)
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

func generateNodeConfig(rootDir string, numVals, numNVals int, noPex bool, startingPort uint64) error {
	// ensure that the directory exists
	// expand the directory path
	rootDir, err := node.ExpandPath(rootDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(rootDir, 0755); err != nil {
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
	// leaderPeerID, err := node.PeerIDFromPubKey(leaderPub)
	leaderPubType := leaderPub.Type()

	genConfig := &config.GenesisConfig{
		ChainID:          "kwil-testnet",
		Leader:           leaderPub.Bytes(), // rethink this so it can be different key types?
		Validators:       make([]ktypes.Validator, numVals),
		DisabledGasCosts: true,
		JoinExpiry:       14400,
		VoteExpiry:       108000,
		MaxBlockSize:     6 * 1024 * 1024,
		MaxVotesPerTx:    200,
	}

	for i := range numVals {
		genConfig.Validators[i] = ktypes.Validator{
			PubKey: keys[i].Public().Bytes(),
			Power:  1,
		}
	}

	// generate the configuration for the nodes
	for i := range numVals + numNVals {
		nodeDir := filepath.Join(rootDir, fmt.Sprintf("node%d", i))
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			return err
		}

		cfg := custom.DefaultConfig() // not config.DefaultConfig(), so custom command config is used

		cfg.PrivateKey = keys[i].Bytes()

		// P2P
		cfg.P2P.Port = startingPort + uint64(i)
		cfg.P2P.IP = "127.0.0.1"
		cfg.P2P.Pex = !noPex

		if i != 0 {
			cfg.P2P.BootNodes = []string{node.FormatPeerString(
				leaderPub.Bytes(), leaderPubType, cfg.P2P.IP, int(startingPort))}
		}

		// DB
		cfg.DB.Port = strconv.Itoa(5432 + i)

		// RPC
		cfg.RPC.ListenAddress = net.JoinHostPort("0.0.0.0", strconv.FormatUint(uint64(8484+i), 10))

		// Admin RPC
		cfg.Admin.ListenAddress = net.JoinHostPort("127.0.0.1", strconv.FormatUint(uint64(8584+i), 10))

		if err := cfg.SaveAs(filepath.Join(nodeDir, config.ConfigFileName)); err != nil {
			return err
		}

		// save the genesis configuration to the root directory
		genFile := filepath.Join(nodeDir, config.GenesisFileName)
		if err := genConfig.SaveAs(genFile); err != nil {
			return err
		}
	}

	return nil
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
