package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/spf13/cobra"
)

var (
	dbidLong = "`" + `dbid` + "`" + ` generates a dbid for a given schema name and deployer.`

	dbidExample = `# Generate a dbid for a schema and deployer
kwil-cli utils dbid --schema=myschema --deployer=0x1234567890abcdef

# Maintain the exact deployer address, and do not trim off the 0x prefix
kwil-cli utils dbid --schema=myschema --deployer=0xnot_an_eth_address --no-trim`
)

func dbidCmd() *cobra.Command {
	var schema, deployer string
	var noTrim bool

	cmd := &cobra.Command{
		Use:     "generate-dbid",
		Short:   dbidLong,
		Long:    dbidLong,
		Example: dbidExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deployerBts := []byte(deployer)
			if strings.HasPrefix(deployer, "0x") && !noTrim {
				deployer = deployer[2:]
				var err error
				deployerBts, err = hex.DecodeString(deployer)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf(`deployer address "%s" has 0x prefix but is not a valid hex string. try using the --no-trim flag`, deployer))
				}
			}

			dbid := utils.GenerateDBID(schema, deployerBts)
			return display.PrintCmd(cmd, &dbidOutput{DBID: dbid})
		},
	}

	cmd.Flags().StringVar(&schema, "schema", "", "Schema name")
	cmd.Flags().StringVar(&deployer, "deployer", "", "Deployer address")
	cmd.Flags().BoolVar(&noTrim, "no-trim", false, "Do not trim off the 0x prefix of the deployer address")
	// mark required flags
	cmd.MarkFlagRequired("schema")
	cmd.MarkFlagRequired("deployer")

	return cmd
}

type dbidOutput struct {
	DBID string `json:"dbid"`
}

func (d *dbidOutput) MarshalJSON() ([]byte, error) {
	// Alias is used to avoid infinite recursion when calling json.Marshal
	type Alias dbidOutput
	return json.Marshal((Alias)(*d))
}

func (d *dbidOutput) MarshalText() (text []byte, err error) {
	return []byte("dbid: " + d.DBID), nil
}
