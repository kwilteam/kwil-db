package setup

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/spf13/cobra"
)

var (
	genesisLong = `` + "`" + `genesis` + "`" + ` creates a new genesis.json file.

This command creates a new genesis file with optionally specified modifications.

If the ` + "`" + `--migration` + "`" + ` flag is set, an incomplete genesis file is generated that can be used in a zero
downtime migration. If generating a migration genesis file, validators and initial state cannot be set.

Validators, balance allocations, and forks should have the format "name:key:power", "address:balance",
and "name:height" respectively.`

	genesisExample = `# Create a new genesis.json file in the current directory
kwil-admin setup genesis

# Create a new genesis.json file in a specific directory with a specific chain ID and a validator with 1 power
kwil-admin setup genesis --out /path/to/directory --chain-id mychainid --validator my_validator:890fe7ae9cb1fa6177555d5651e1b8451b4a9c64021c876236c700bc2690ff1d:1

# Create a new genesis.json with the specified allocation
kwil-admin setup genesis --alloc 0x7f5f4552091a69125d5dfcb7b8c2659029395bdf:100

# Create a new genesis.json file to be used in a network migration
kwil-admin setup genesis --migration --out /path/to/directory --chain-id mychainid`
)

func genesisCmd() *cobra.Command {
	var validators, allocs, forks []string
	var chainID, output, genesisState string
	var migration, withGasCosts bool
	var maxBytesPerBlock, joinExpiry, voteExpiry, maxVotesPerBlock int64
	cmd := &cobra.Command{
		Use:     "genesis",
		Short:   "`genesis` creates a new genesis.json file",
		Long:    genesisLong,
		Example: genesisExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := common.ExpandPath(output)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to expand output path: %w", err))
			}

			err = os.MkdirAll(output, 0755)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to create output directory: %w", err))
			}

			out := filepath.Join(output, "genesis.json")

			makeErr := func(e error) error {
				return display.PrintErr(cmd, fmt.Errorf("failed to create genesis file: %w", e))
			}

			genesisInfo := chain.DefaultGenesisConfig()
			if cmd.Flags().Changed(validatorsFlag) {
				if migration {
					return makeErr(errors.New("cannot set validators when generating a migration genesis file"))
				}

				for _, v := range validators {
					parts := strings.Split(v, ":")
					if len(parts) != 3 {
						return makeErr(errors.New("validator must have the format 'name:key:power', received: " + v))
					}

					hexPub, err := hex.DecodeString(parts[1])
					if err != nil {
						return makeErr(fmt.Errorf("failed to decode hex public key '%s': %w", parts[1], err))
					}

					power, err := strconv.ParseInt(parts[2], 10, 64)
					if err != nil {
						return makeErr(fmt.Errorf("failed to parse power '%s': %w", parts[2], err))
					}

					genesisInfo.Validators = append(genesisInfo.Validators, &chain.GenesisValidator{
						PubKey: hexPub,
						Power:  power,
						Name:   parts[0],
					})
				}
			}
			if cmd.Flags().Changed(allocsFlag) {
				for _, a := range allocs {
					parts, err := splitDelimitedStrings(a)
					if err != nil {
						return makeErr(err)
					}

					balance, ok := new(big.Int).SetString(parts[1], 10)
					if !ok {
						return makeErr(fmt.Errorf("failed to parse balance '%s'", parts[1]))
					}

					genesisInfo.Alloc[parts[0]] = balance
				}
			}
			if cmd.Flags().Changed(forksFlag) {
				genesisInfo.ForkHeights = make(map[string]*uint64)
				for _, f := range forks {
					parts := strings.Split(f, ":")
					if len(parts) != 2 {
						return makeErr(errors.New("fork must have the format 'height:name', received: " + f))
					}

					height, err := strconv.ParseInt(parts[1], 10, 64)
					if err != nil {
						return makeErr(fmt.Errorf("failed to parse height '%s': %w", parts[0], err))
					}

					if height < 0 {
						// skip negative heights
						continue
					}

					uintHeight := uint64(height)
					genesisInfo.ForkHeights[parts[1]] = &uintHeight
				}
			}
			if cmd.Flags().Changed(chainIDFlag) {
				if migration {
					return makeErr(errors.New("cannot set chain ID when generating a migration genesis file"))
				}

				genesisInfo.ChainID = chainID
			}
			if cmd.Flags().Changed(genesisStateFlag) {
				if migration {
					return makeErr(errors.New("cannot set genesis state when generating a migration genesis file"))
				}

				apphash, err := appHashFromSnapshotFile(genesisState)
				if err != nil {
					return makeErr(err)
				}

				genesisInfo.DataAppHash = apphash
			}
			genesisInfo.ConsensusParams.WithoutGasCosts = !withGasCosts
			if cmd.Flags().Changed(maxBytesPerBlockFlag) {
				genesisInfo.ConsensusParams.Block.MaxBytes = maxBytesPerBlock
			}
			if cmd.Flags().Changed(joinExpiryFlag) {
				genesisInfo.ConsensusParams.Validator.JoinExpiry = joinExpiry
			}
			if cmd.Flags().Changed(voteExpiryFlag) {
				genesisInfo.ConsensusParams.Votes.VoteExpiry = voteExpiry
			}
			if cmd.Flags().Changed(maxVotesPerTxFlag) {
				genesisInfo.ConsensusParams.Votes.MaxVotesPerTx = maxVotesPerBlock
			}

			if migration {
				genesisInfo.Validators = nil
				genesisInfo.Alloc = nil
				genesisInfo.ForkHeights = nil
			}

			existingFile, err := os.Stat(out)
			if err == nil && existingFile.IsDir() {
				return makeErr(fmt.Errorf("a directory already exists at %s, please remove it first", out))
			} else if err == nil {
				return makeErr(fmt.Errorf("%s already exists, please remove it first", out))
			}

			err = genesisInfo.SaveAs(out)
			if err != nil {
				return makeErr(err)
			}

			return display.PrintCmd(cmd, display.RespString("Created genesis.json file at "+out))
		},
	}

	cmd.Flags().StringVar(&output, outFlag, "", "Output directory for the genesis.json file")
	cmd.Flags().StringVar(&chainID, chainIDFlag, "", "Chain ID for the genesis.json file")
	cmd.Flags().StringArrayVar(&validators, validatorsFlag, nil, "Public keys and power of initial validator(s)")
	cmd.Flags().StringArrayVar(&allocs, allocsFlag, nil, "Address and initial balance allocation(s)")
	cmd.Flags().StringArrayVar(&forks, forksFlag, nil, "Block height and name of fork(s)")
	cmd.Flags().BoolVar(&withGasCosts, withGasCostsFlag, false, "Include gas costs in the genesis file")
	cmd.Flags().Int64Var(&maxBytesPerBlock, maxBytesPerBlockFlag, 0, "Maximum number of bytes per block")
	cmd.Flags().Int64Var(&joinExpiry, joinExpiryFlag, 0, "Number of blocks before a join proposal expires")
	cmd.Flags().Int64Var(&voteExpiry, voteExpiryFlag, 0, "Number of blocks before a vote proposal expires")
	cmd.Flags().Int64Var(&maxVotesPerBlock, maxVotesPerTxFlag, 0, "Maximum number of votes per validator transaction (each validator has 1 validator tx per block)")
	cmd.Flags().BoolVar(&migration, migrationFlag, false, "Generate an incomplete genesis file for zero downtime migration")
	cmd.Flags().StringVar(&genesisState, genesisStateFlag, "", "Path to a genesis state file")

	return cmd
}

const (
	outFlag              = "out"
	chainIDFlag          = "chain-id"
	validatorsFlag       = "validator"
	allocsFlag           = "alloc"
	forksFlag            = "fork"
	withGasCostsFlag     = "with-gas-costs"
	maxBytesPerBlockFlag = "max-bytes-per-block"
	joinExpiryFlag       = "join-expiry"
	voteExpiryFlag       = "vote-expiry"
	maxVotesPerTxFlag    = "max-votes-per-tx"
	migrationFlag        = "migration"
	genesisStateFlag     = "genesis-state"
)

// splitDelimitedStrings splits a string into two parts using a colon as the delimiter.
// It returns an error if the string does not contain exactly one colon.
func splitDelimitedStrings(s string) ([2]string, error) {
	if strings.Count(s, ":") != 1 {
		return [2]string{}, errors.New("must only have one delimiting colon, received: " + s)
	}

	parts := strings.Split(s, ":")
	return [2]string{parts[0], parts[1]}, nil
}
