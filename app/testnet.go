package app

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	mrand2 "math/rand/v2"
	"os"
	"path/filepath"

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
			rootDir, err := RootDir(cmd)
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
		rr := mrand2.NewChaCha8(seed)
		priv := node.NewKey(rr)
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

		cfg := config.DefaultConfig()
		cfg.PrivateKey = privKey
		cfg.P2P.Port = startingPort + uint64(i)
		cfg.P2P.IP = "127.0.0.1"
		cfg.P2P.Pex = !noPex

		if i != 0 {
			cfg.P2P.BootNodes = []string{node.FormatPeerString(
				leaderPub.Bytes(), leaderPubType, cfg.P2P.IP, int(startingPort))}
		}

		if err := cfg.SaveAs(filepath.Join(nodeDir, ConfigFileName)); err != nil {
			return err
		}

		// save the genesis configuration to the root directory
		genFile := filepath.Join(nodeDir, GenesisFileName)
		if err := genConfig.SaveAs(genFile); err != nil {
			return err
		}
	}

	return nil
}
