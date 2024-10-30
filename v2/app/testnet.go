package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	mrand2 "math/rand/v2"
	"os"
	"path/filepath"

	"kwil/node"
	"kwil/node/types"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

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

	if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
		return err
	}

	var keys []crypto.PrivKey
	// generate the configuration for the nodes
	for i := 0; i < numVals+numNVals; i++ {
		// generate Keys, so that the connection strings and the validator set can be generated before the node config files are generated
		rr := rand.Reader
		var seed [32]byte
		binary.LittleEndian.PutUint64(seed[:], startingPort+uint64(i))
		seed = sha256.Sum256(seed[:])
		rr = mrand2.NewChaCha8(seed)
		priv := node.NewKey(rr)
		keys = append(keys, priv)
	}

	leaderR, err := keys[0].GetPublic().Raw()
	if err != nil {
		return err
	}
	leaderAddr, err := peer.IDFromPrivateKey(keys[0])
	if err != nil {
		return err
	}

	genConfig := &node.GenesisConfig{
		Leader:     leaderR,
		Validators: make([]types.Validator, numVals),
	}

	for i := 0; i < numVals; i++ {
		pub, err := keys[i].GetPublic().Raw()
		if err != nil {
			return err
		}
		genConfig.Validators[i] = types.Validator{
			PubKey: pub,
			Power:  1,
		}
	}

	// generate the configuration for the nodes
	for i := 0; i < numVals+numNVals; i++ {
		nodeDir := filepath.Join(rootDir, fmt.Sprintf("node%d", i))
		if err := os.MkdirAll(nodeDir, os.ModePerm); err != nil {
			return err
		}

		privKey, err := keys[i].Raw()
		if err != nil {
			return err
		}

		nodeConfig := &node.NodeConfig{
			Port:       startingPort + uint64(i),
			IP:         "127.0.0.1",
			SeedNode:   "",
			Pex:        !noPex,
			PrivateKey: privKey,
		}

		if i != 0 {
			nodeConfig.SeedNode = fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", nodeConfig.IP, startingPort, leaderAddr)
		}

		if err := nodeConfig.SaveAs(filepath.Join(nodeDir, "config.json")); err != nil {
			return err
		}

		// save the genesis configuration to the root directory
		genFile := filepath.Join(nodeDir, "genesis.json")
		if err := genConfig.SaveAs(genFile); err != nil {
			return err
		}
	}

	return nil
}
