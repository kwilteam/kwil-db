package setup

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/key"
	"github.com/kwilteam/kwil-db/app/node/conf"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/node"
	"github.com/spf13/cobra"
)

const (
	nodeDirPerm = 0755
)

var (
	genesisValidatorGas, _ = big.NewInt(0).SetString("10000000000000000000000", 10)
)

var (
	initLong = `The init command facilitates quick setup of an isolated Kwil node either on a fresh network in which that node is the single validator or to join an existing network.
This permits rapid prototyping and evaluation of Kwil functionality. An output directory can be specified using the ` + "`" + `--output-dir` + "`" + `" flag.
If no output directory is specified, the node will be initialized ` + "`" + `./testnet` + "`" + `.`

	initExample = `# Initialize a node, with a new network, in the directory ~/.kwil-new
kwild setup init -r ~/.kwild-new`
)

func InitCmd() *cobra.Command {
	var genesisPath, genesisState string
	var genFlags genesisFlagConfig

	cmd := &cobra.Command{
		Use:               "init",
		Short:             "Generate configuration for a Kwil node.",
		Long:              initLong,
		Example:           initExample,
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			genSnapshotFlag := cmd.Flag("genesis-snapshot").Changed
			genStateFlag := cmd.Flag("genesis-state").Changed
			if genSnapshotFlag && genStateFlag {
				return display.PrintErr(cmd, errors.New("cannot use both --genesis-snapshot and --genesis-state, use just genesis-snapshot flag instead"))
			}

			rootDir, err := bind.RootDir(cmd)
			if err != nil {
				return err
			}

			// Ensure the root directory exists
			outDir, err := node.ExpandPath(rootDir)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(outDir, nodeDirPerm); err != nil {
				return err
			}

			cfg := conf.ActiveConfig()

			bind.Debugf("effective node config (toml):\n%s", bind.LazyPrinter(func() string {
				rawToml, err := cfg.ToTOML()
				if err != nil {
					return fmt.Errorf("failed to marshal config to toml: %w", err).Error()
				}
				return string(rawToml)
			}))

			if cmd.Flags().Changed("genesis-snapshot") {
				genesisState = genFlags.genesisState
				genesisState, err = node.ExpandPath(genesisState)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				stateFile := config.GenesisStateFileName(outDir)

				if err := utils.CopyFile(genesisState, stateFile); err != nil {
					return display.PrintErr(cmd, err)
				}
				cfg.GenesisState = stateFile
			}

			// Generate and save the node key to the root directory
			privKey, err := crypto.GeneratePrivateKey(crypto.KeyTypeSecp256k1)
			if err != nil {
				return err
			}

			if err := key.SaveNodeKey(config.NodeKeyFilePath(rootDir), privKey); err != nil {
				return err
			}

			var genCfg *config.GenesisConfig
			genFile := filepath.Join(outDir, "genesis.json")
			if genesisPath != "" { // Init for the node to join an existing network
				if genesisPath, err = node.ExpandPath(genesisPath); err != nil {
					return display.PrintErr(cmd, err)
				}

				if err := utils.CopyFile(genesisPath, genFile); err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to copy genesis file: %w", err))
				}
			} else { // Init command for creating a new network, new genesis file will be created
				genCfg = config.DefaultGenesisConfig()
				genCfg, err = mergeGenesisFlags(genCfg, cmd, &genFlags)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				genCfg.Leader = types.PublicKey{PublicKey: privKey.Public()}
				genCfg.Validators = append(genCfg.Validators, &types.Validator{
					AccountID: types.AccountID{
						Identifier: privKey.Public().Bytes(),
						KeyType:    privKey.Type(),
					},
					Power: 1,
				})

				// If DB owner is not set, set it to the node's public key
				if genCfg.DBOwner == "" {
					signer := auth.GetUserSigner(privKey)
					ident, err := auth.GetIdentifierFromSigner(signer)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("failed to get identifier for dbOwner: %w", err))
					}
					genCfg.DBOwner = ident
				}

				// allocate some initial balance to validators if gas is enabled and
				// if no funds are allocated to that validators.
				if !genCfg.DisabledGasCosts {
					for _, v := range genCfg.Validators {
						hexPubKey := v.Identifier.String()
						genCfg.Allocs = append(genCfg.Allocs, config.GenesisAlloc{
							ID:      hexPubKey,
							KeyType: v.Identifier.String(),
							Amount:  genesisValidatorGas,
						})

					}
				}

				if err := genCfg.SaveAs(genFile); err != nil {
					return display.PrintErr(cmd, err)
				}
			}

			// Save the config to the root directory
			if err := cfg.SaveAs(config.ConfigFilePath(rootDir)); err != nil {
				return err
			}

			return nil
		},
	}

	defaultCfg := custom.DefaultConfig()
	bind.SetFlagsFromStruct(cmd.Flags(), defaultCfg)
	bindGenesisFlags(cmd, &genFlags)

	// genesis config flags
	cmd.Flags().StringVarP(&genesisPath, "genesis", "g", "", "path to genesis file")

	return cmd
}
