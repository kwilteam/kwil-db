package cli

import (
	"fmt"
	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group kwil queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListDatabases())
	cmd.AddCommand(CmdShowDatabases())
	cmd.AddCommand(CmdListDdl())
	cmd.AddCommand(CmdShowDdl())
	cmd.AddCommand(CmdListDdlindex())
	cmd.AddCommand(CmdShowDdlindex())
	cmd.AddCommand(CmdListQueryids())
	cmd.AddCommand(CmdShowQueryids())
	// this line is used by starport scaffolding # 1

	return cmd
}
