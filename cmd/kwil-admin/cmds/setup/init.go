package setup

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/spf13/cobra"
)

var (
	initLong = `The ` + "`" + `init` + "`" + ` command facilitates quick setup of an isolated Kwil node on a fresh network in which that node is the single validator.
This permits rapid prototyping and evaluation of Kwil functionality. An output directory can be specified using the ` + "`" + `--output-dir` + "`" + `" flag.
If no output directory is specified, the node will be initialized ` + "`" + `./testnet` + "`" + `.`

	initExample = `# Initialize a node, with a new network, in the directory ~/.kwil-new
kwil-admin setup init -o ~/.kwild-new`
)

func initCmd() *cobra.Command {
	var out, chainId, genesisPath string
	var blockInterval time.Duration
	var joinExpiry int64 // block height
	var withGas bool
	var allocs AllocsFlag
	cfg := cmd.DefaultConfig()

	cmd1 := &cobra.Command{
		Use:     "init",
		Short:   "The `init` command facilitates quick setup of an isolated Kwil node.",
		Long:    initLong,
		Example: initExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			outFlag := cmd.Flag("output-dir").Changed
			rootFlag := cmd.Flag("root-dir").Changed

			if outFlag && rootFlag {
				return display.PrintErr(cmd, errors.New("cannot use both --output-dir and --root-dir"))
			}

			// if --root-dir is set, use that as the output directory, over the deprecated --output-dir defaults
			if rootFlag {
				out = cfg.RootDir
			}

			expandedDir, err := common.ExpandPath(out)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// Create the output directory
			if err = os.MkdirAll(expandedDir, 0755); err != nil {
				return display.PrintErr(cmd, err)
			}
			cfg.RootDir = expandedDir

			// saves config and private key files in the root directory
			pub, err := nodecfg.GenerateNodeFiles(expandedDir, cfg, true)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			genFile := filepath.Join(expandedDir, cometbft.GenesisJSONName)
			if genesisPath != "" {
				if genesisPath, err = common.ExpandPath(genesisPath); err != nil {
					return display.PrintErr(cmd, err)
				}

				file, err := os.ReadFile(genesisPath)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				err = os.WriteFile(genFile, file, 0644)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
			} else {
				genCfg := &nodecfg.NodeGenerateConfig{
					ChainID:         chainId,
					BlockInterval:   blockInterval,
					OutputDir:       expandedDir,
					JoinExpiry:      joinExpiry,
					WithoutGasCosts: !withGas,
					Allocs:          allocs.M,
				}

				_, err = os.Stat(genFile)
				if os.IsNotExist(err) {
					genesisCfg := chain.NewGenesisWithValidator(pub)
					genCfg.ApplyGenesisParams(genesisCfg)
					if err = genesisCfg.SaveAs(genFile); err != nil {
						return display.PrintErr(cmd, err)
					}
				}
			}
			return display.PrintCmd(cmd, display.RespString("Initialized node in "+expandedDir))
		},
	}

	// genesis.json flags
	cmd1.Flags().StringVarP(&genesisPath, "genesis", "g", "", "path to genesis file")
	cmd1.Flags().StringVar(&chainId, "chain-id", "", "chain ID to use for the genesis file")
	cmd1.Flags().Int64Var(&joinExpiry, "join-expiry", 14400, "number of blocks before a join request expires")
	cmd1.Flags().BoolVar(&withGas, "gas", false, "enable gas")
	cmd1.Flags().Var(&allocs, "alloc", "account=amount pairs of genesis account allocations")

	// config.toml flags
	config.AddConfigFlags(cmd1.Flags(), cfg)

	// TODO: deprecate below flags in v0.10.0
	cmd1.Flags().StringVarP(&out, "output-dir", "o", "./.testnet", "generated node parent directory. To be deprecated in v0.10.0, until then --root-dir is ignored")
	err := cmd1.Flags().MarkDeprecated("output-dir", "use --cfg.root-dir instead from v0.10.0")
	if err != nil {
		panic(err)
	}

	cmd1.Flags().DurationVarP(&blockInterval, "block-interval", "i", 6*time.Second, "shortest block interval in seconds. To be deprecated in v0.10.0")
	err = cmd1.Flags().MarkDeprecated("block-interval", "use --chain.consensus.timeout-commit instead from v0.10.0")
	if err != nil {
		panic(err)
	}

	return cmd1
}

type AllocsFlag struct {
	M map[string]*big.Int
}

func (a *AllocsFlag) String() string {
	return fmt.Sprintf("%v", a.M)
}

func (a *AllocsFlag) Set(value string) error {
	if a.M == nil {
		a.M = map[string]*big.Int{}
	}
	split := strings.Split(value, "=")
	if len(split) != 2 {
		return errors.New("invalid format for alloc, expected key=value")
	}
	amt, ok := big.NewInt(0).SetString(split[1], 10)
	if !ok {
		return errors.New("bad amount")
	}
	a.M[split[0]] = amt
	return nil
}

func (a *AllocsFlag) Type() string {
	return "allocFlag"
}
