package setup

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

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

	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Generate configuration for multiple nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := bind.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}
			return generateNodeConfig(rootDir, numVals, numNVals, noPex, startingPort)
		},
	}

	cmd.Flags().IntVarP(&numVals, "vals", "v", 3, "number of validators (includes the one leader)")
	cmd.Flags().IntVarP(&numNVals, "non-vals", "n", 0, "number of non-validators")
	cmd.Flags().BoolVar(&noPex, "no-pex", false, "disable peer exchange")
	cmd.Flags().Uint64VarP(&startingPort, "port", "p", 6600, "starting port for the nodes")

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
		Leader:     leaderPub.Bytes(), // rethink this so it can be different key types?
		Validators: make([]ktypes.Validator, numVals),
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

		privKey := keys[i].Bytes()

		cfg := custom.DefaultConfig() // not config.DefaultConfig(), so custom command config is used
		cfg.PrivateKey = privKey
		cfg.P2P.Port = startingPort + uint64(i)
		cfg.P2P.IP = "127.0.0.1"
		cfg.P2P.Pex = !noPex

		if i != 0 {
			cfg.P2P.BootNodes = []string{node.FormatPeerString(
				leaderPub.Bytes(), leaderPubType, cfg.P2P.IP, int(startingPort))}
		}

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
