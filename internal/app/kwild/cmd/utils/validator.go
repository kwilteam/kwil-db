package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
)

// GenValidatorCmd allows the generation of a keypair for a
// validator.

func GenValidatorCmd() *cobra.Command {
	validatorCmd := &cobra.Command{
		Use:     "gen-validator",
		Aliases: []string{"gen_validator"},
		Short:   "Generate new validator keypair",
		Run:     genValidator,
	}
	validatorCmd.Flags().StringVar(&outputDir, "o", "./mytestnet", "directory to store initialization data for the testnet")
	return validatorCmd
}

func genValidator(cmd *cobra.Command, args []string) {
	pv := privval.GenFilePV("", "")
	jsbz, err := cmtjson.Marshal(pv)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filepath.Join(outputDir, "config/priv_validator_key.json"), jsbz, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(`%v`, string(jsbz))
}
