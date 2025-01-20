package setup

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node"
)

var (
	genesisLong = `The ` + "`genesis`" + ` command creates a new ` + "`genesis.json`" + ` file with optionally specified modifications.

Validators and balance allocations should have the format "pubkey:power", "address:balance" respectively.`

	genesisExample = `# Create a new genesis.json file in the current directory
kwild setup genesis

# Create a new genesis.json file in a specific directory with a specific chain ID and a validator with 1 power
kwild setup genesis --out /path/to/directory --chain-id mychainid --validator 890fe7ae9cb1fa6177555d5651e1b8451b4a9c64021c876236c700bc2690ff1d:1

# Create a new genesis.json with the specified allocation
kwild setup genesis --alloc 0x7f5f4552091a69125d5dfcb7b8c2659029395bdf:100`
)

type genesisFlagConfig struct {
	chainID       string
	validators    []string
	allocs        []string
	withGas       bool
	leader        string
	dbOwner       string
	maxBlockSize  int64
	joinExpiry    time.Duration
	maxVotesPerTx int64
	genesisState  string
}

func GenesisCmd() *cobra.Command {
	var flagCfg genesisFlagConfig
	var output string

	cmd := &cobra.Command{
		Use:               "genesis",
		Short:             "Create a new genesis.json file",
		Long:              genesisLong,
		Example:           genesisExample,
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir, err := node.ExpandPath(output)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to expand output path: %w", err))
			}

			err = os.MkdirAll(outDir, nodeDirPerm)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to create output directory: %w", err))
			}

			genesisFile := config.GenesisFilePath(outDir)

			conf := config.DefaultGenesisConfig()
			conf, err = mergeGenesisFlags(conf, cmd, &flagCfg)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to create genesis file: %w", err))
			}

			existingFile, err := os.Stat(genesisFile)
			if err == nil && existingFile.IsDir() {
				return display.PrintErr(cmd, fmt.Errorf("a directory already exists at %s, please remove it first", genesisFile))
			} else if err == nil {
				return display.PrintErr(cmd, fmt.Errorf("file already exists at %s, please remove it first", genesisFile))
			}

			err = conf.SaveAs(genesisFile)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to save genesis file: %w", err))
			}

			return display.PrintCmd(cmd, display.RespString("Created genesis.json file at "+genesisFile))
		},
	}

	bindGenesisFlags(cmd, &flagCfg)
	cmd.Flags().StringVar(&output, "out", "", "Output directory for the genesis.json file")

	return cmd
}

// bindGenesisFlags binds the genesis configuration flags to the given command.
func bindGenesisFlags(cmd *cobra.Command, cfg *genesisFlagConfig) {
	cmd.Flags().StringVar(&cfg.chainID, "chain-id", "", "chainID for the genesis.json file")
	cmd.Flags().StringSliceVar(&cfg.validators, "validators", nil, "public key, keyType and power of initial validator(s)") // accept: [hexpubkey1#keyType1:power1]
	cmd.Flags().StringSliceVar(&cfg.allocs, "allocs", nil, "address and initial balance allocation(s)")
	cmd.Flags().BoolVar(&cfg.withGas, "with-gas", false, "include gas costs in the genesis.json file")
	cmd.Flags().StringVar(&cfg.leader, "leader", "", "public key of the block proposer")
	cmd.Flags().StringVar(&cfg.dbOwner, "db-owner", "", "owner of the database")
	cmd.Flags().Int64Var(&cfg.maxBlockSize, "max-block-size", 0, "maximum block size")
	cmd.Flags().DurationVar(&cfg.joinExpiry, "join-expiry", 0, "Number of blocks before a join proposal expires")
	cmd.Flags().Int64Var(&cfg.maxVotesPerTx, "max-votes-per-tx", 0, "Maximum votes per transaction")
	cmd.Flags().StringVar(&cfg.genesisState, "genesis-snapshot", "", "path to genesis state snapshot file")
}

// mergeGenesisFlags merges the genesis configuration flags with the given configuration.
func mergeGenesisFlags(conf *config.GenesisConfig, cmd *cobra.Command, flagCfg *genesisFlagConfig) (*config.GenesisConfig, error) {
	if cmd.Flags().Changed("chain-id") {
		conf.ChainID = flagCfg.chainID
	}

	if cmd.Flags().Changed("validators") {
		conf.Validators = nil
		for _, v := range flagCfg.validators {
			parts := strings.Split(v, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid format for validator, expected key:power, received: %s", v)
			}

			keyParts := strings.Split(parts[0], "#")
			hexPub, err := hex.DecodeString(keyParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid public key for validator: %s", parts[0])
			}

			power, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid power for validator: %s", parts[1])
			}

			keyType, err := crypto.ParseKeyType(keyParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid key type for validator: %s", keyParts[1])
			}

			conf.Validators = append(conf.Validators, &types.Validator{
				AccountID: types.AccountID{
					Identifier: hexPub,
					KeyType:    keyType,
				},
				Power: power,
			})
		}
	}

	if cmd.Flags().Changed("allocs") {
		conf.Allocs = nil
		for _, a := range flagCfg.allocs {
			parts := strings.Split(a, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid format for alloc, expected id#keyType:balance, received: %s", a)
			}

			keyParts := strings.Split(parts[0], "#")
			if len(keyParts) != 2 {
				return nil, fmt.Errorf("invalid address for alloc: %s", parts[0])
			}

			keyType, err := crypto.ParseKeyType(keyParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid key type for validator: %s", keyParts[1])
			}

			balance, ok := new(big.Int).SetString(parts[1], 10)
			if !ok {
				return nil, fmt.Errorf("invalid balance for alloc: %s", parts[1])
			}

			keyStr, err := hex.DecodeString(keyParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid address for alloc: %s", keyParts[0])
			}

			conf.Allocs = append(conf.Allocs, config.GenesisAlloc{
				ID: config.KeyHexBytes{
					HexBytes: keyStr,
				},
				KeyType: keyType.String(),
				Amount:  balance,
			})
		}
	}

	if cmd.Flags().Changed("with-gas") {
		conf.DisabledGasCosts = !flagCfg.withGas
	}

	if cmd.Flags().Changed("leader") {
		pubkeyBts, keyType, err := config.DecodePubKeyAndType(flagCfg.leader)
		if err != nil {
			return nil, err
		}
		pubkey, err := crypto.UnmarshalPublicKey(pubkeyBts, keyType)
		if err != nil {
			return nil, err
		}
		conf.Leader = types.PublicKey{PublicKey: pubkey}
	}

	if cmd.Flags().Changed("db-owner") {
		conf.DBOwner = flagCfg.dbOwner
	}

	if cmd.Flags().Changed("max-block-size") {
		conf.MaxBlockSize = flagCfg.maxBlockSize
	}

	if cmd.Flags().Changed("join-expiry") {
		conf.JoinExpiry = types.Duration(flagCfg.joinExpiry)
	}

	if cmd.Flags().Changed("max-votes-per-tx") {
		conf.MaxVotesPerTx = flagCfg.maxVotesPerTx
	}

	if cmd.Flags().Changed("genesis-state") {
		hash, err := appHashFromSnapshotFile(flagCfg.genesisState)
		if err != nil {
			return nil, err
		}
		conf.StateHash = hash
	}

	return conf, nil
}
