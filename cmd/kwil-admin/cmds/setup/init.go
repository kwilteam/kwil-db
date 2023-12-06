package setup

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
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
	var out, chainId string
	var blockInterval time.Duration
	var joinExpiry int64 // block height
	var withoutNonces, withGas bool
	var allocs AllocsFlag

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "The `init` command facilitates quick setup of an isolated Kwil node.",
		Long:    initLong,
		Example: initExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			expandedDir, err := expandPath(out)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			genCfg := &nodecfg.NodeGenerateConfig{
				ChainID:         chainId,
				BlockInterval:   blockInterval,
				OutputDir:       expandedDir,
				JoinExpiry:      joinExpiry,
				WithoutGasCosts: !withGas,
				WithoutNonces:   withoutNonces,
				Allocs:          allocs.M,
			}

			// GenerateNodeConfig fmt.Printlns, but do we want this printed to display pkg?
			err = nodecfg.GenerateNodeConfig(genCfg)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespString("Initialized node in "+expandedDir))
		},
	}

	cmd.Flags().StringVarP(&out, "output-dir", "o", "./.testnet", "generated node parent directory [default: ./testnet]")
	cmd.Flags().StringVar(&chainId, "chain-id", "", "chain ID to use for the genesis file (default: random)")
	cmd.Flags().DurationVarP(&blockInterval, "block-interval", "i", 6*time.Second, "shortest block interval in seconds (timeout_commit) [default: 6s]")
	cmd.Flags().Int64Var(&joinExpiry, "join-expiry", 14400, "number of blocks before a join request expires [default: 14400]")
	cmd.Flags().BoolVar(&withoutNonces, "without-nonces", false, "disable account nonces")
	cmd.Flags().BoolVar(&withGas, "gas", false, "enable gas")
	cmd.Flags().Var(&allocs, "alloc", "account=amount pairs of genesis account allocations")

	return cmd
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
