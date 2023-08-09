package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

// InitFilesCmd initializes a fresh CometBFT instance.
func InitFilesCmd() *cobra.Command {
	var initFilesCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes files required for a kwil node",
		RunE:  initFiles,
	}
	initFilesCmd.Flags().StringVar(&outputDir, "home", "", "comet home directory")
	if outputDir == "" {
		outputDir = os.Getenv("COMET_BFT_HOME")
	}
	initFilesCmd.Flags().BoolVar(&disable_gas, "disable-gas", false,
		"Disables gas costs on all transactions and once the network is initialized, it can't be changed")
	return initFilesCmd
}
func initFiles(cmd *cobra.Command, args []string) error {
	config := cfg.DefaultConfig()
	config.SetRoot(outputDir)
	err := os.MkdirAll(filepath.Join(outputDir, "config"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(outputDir)
		return err
	}

	err = os.MkdirAll(filepath.Join(outputDir, "data"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(outputDir)
		return err
	}
	err = InitFilesWithConfig(config)
	if err != nil {
		return err
	}
	pvKeyFile := filepath.Join(outputDir, config.BaseConfig.PrivValidatorKey)
	pvStateFile := filepath.Join(outputDir, config.BaseConfig.PrivValidatorState)
	pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

	pubKey, err := pv.GetPubKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	genVal := types.GenesisValidator{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		Power:   1,
		Name:    "node-0",
	}

	var chainID string
	if disable_gas {
		chainID = "kwil-chain-gcd-"
	} else {
		chainID = "kwil-chain-gce-"
	}

	vals := []types.GenesisValidator{genVal}

	genDoc := &types.GenesisDoc{
		ChainID:         chainID + cmtrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     cmttime.Now(),
		InitialHeight:   initialHeight,
		Validators:      vals,
	}

	if err := genDoc.SaveAs(filepath.Join(outputDir, config.BaseConfig.Genesis)); err != nil {
		_ = os.RemoveAll(outputDir)
		return err
	}

	config.P2P.AddrBookStrict = false
	config.P2P.AllowDuplicateIP = true
	config.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	cfg.WriteConfigFile(filepath.Join(outputDir, "config", "config.toml"), config)
	return nil
}
