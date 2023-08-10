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

var initFlags = struct {
	initialHeight int64
	homeDir       string
}{}

// InitFilesCmd initializes a fresh CometBFT instance.
func InitFilesCmd() *cobra.Command {
	var initFilesCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes files required for a kwil node",
		RunE:  initFiles,
	}
	initFilesCmd.Flags().StringVar(&initFlags.homeDir, "home", "", "comet home directory")
	if initFlags.homeDir == "" {
		initFlags.homeDir = os.Getenv("COMET_BFT_HOME")
	}
	initFilesCmd.Flags().Int64Var(&initFlags.initialHeight, "initial-height", 0, "initial height of the first block")

	return initFilesCmd
}
func initFiles(cmd *cobra.Command, args []string) error {
	config := cfg.DefaultConfig()
	config.SetRoot(initFlags.homeDir)
	err := os.MkdirAll(filepath.Join(initFlags.homeDir, "config"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(initFlags.homeDir)
		return err
	}

	err = os.MkdirAll(filepath.Join(initFlags.homeDir, "data"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(initFlags.homeDir)
		return err
	}
	err = InitFilesWithConfig(config)
	if err != nil {
		return err
	}
	pvKeyFile := filepath.Join(initFlags.homeDir, config.BaseConfig.PrivValidatorKey)
	pvStateFile := filepath.Join(initFlags.homeDir, config.BaseConfig.PrivValidatorState)
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

	chainIDPrefix := "kwil-chain-"

	vals := []types.GenesisValidator{genVal}

	genDoc := &types.GenesisDoc{
		ChainID:         chainIDPrefix + cmtrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     cmttime.Now(),
		InitialHeight:   initFlags.initialHeight,
		Validators:      vals,
	}

	if err := genDoc.SaveAs(filepath.Join(initFlags.homeDir, config.BaseConfig.Genesis)); err != nil {
		_ = os.RemoveAll(initFlags.homeDir)
		return err
	}

	config.P2P.AddrBookStrict = false
	config.P2P.AllowDuplicateIP = true
	config.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	cfg.WriteConfigFile(filepath.Join(initFlags.homeDir, "config", "config.toml"), config)
	return nil
}
