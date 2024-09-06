// package custom allows for the creation of a custom-branded CLI that packages together the kwil-cli, kwil-admin, and kwild CLIs.
package custom

import (
	"fmt"

	"github.com/spf13/cobra"

	kwilAdminRoot "github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds"
	kwilCLIRoot "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds"
	kwildConfig "github.com/kwilteam/kwil-db/cmd/kwild/config"
	kwildRoot "github.com/kwilteam/kwil-db/cmd/kwild/root"
	"github.com/kwilteam/kwil-db/common/config"
)

// CommonCmdConfig configures the root command.
type CommonCmdConfig struct {
	// RootCmd is the name of the command.
	RootCmd string
	// ProjectName is the name of the project, which will be used in the help text.
	ProjectName string
	// DefaultConfig is a function that allows the default configuration to be changed.
	DefaultConfig func(*config.KwildConfig)
}

var longDesc = `Command line interface client for using %s.

There are 3 subcommands:

- ` + "`" + `node` + "`" + `: Runs the %s node and RPC server.
- ` + "`" + `client` + "`" + `: Command line interface for an RPC client.
- ` + "`" + `admin` + "`" + `: Utilities for managing a %s node and network participation.

For guides and reference documentation, see the following links. The links document ` + "`" + `kwild` + "`" + `,
` + "`" + `kwil-cli` + "`" + `, and ` + "`" + `kwil-admin` + "`" + `, which directly correspond to the ` + "`" + `node` + "`" +
	`, ` + "`" + `client` + "`" + `, and ` + "`" + `admin` + "`" + `
subcommands, respectively:

- Node: https://docs.kwil.com/docs/node/quickstart
- Client: https://docs.kwil.com/docs/ref/kwil-cli
- Admin: https://docs.kwil.com/docs/admin/installation`

func NewCustomCmd(cmdConfig CommonCmdConfig) *cobra.Command {
	root := &cobra.Command{
		Use:   cmdConfig.RootCmd,
		Short: "Command line interface client for using " + cmdConfig.ProjectName + ".",
		Long:  fmt.Sprintf(longDesc, cmdConfig.ProjectName, cmdConfig.ProjectName, cmdConfig.ProjectName),
	}

	if cmdConfig.DefaultConfig == nil {
		cmdConfig.DefaultConfig = func(c *config.KwildConfig) {}
	}

	if cmdConfig.ProjectName == "" {
		cmdConfig.ProjectName = "kwil"
	}

	if cmdConfig.RootCmd == "" {
		cmdConfig.RootCmd = cmdConfig.ProjectName
	}

	oldDefault := kwildConfig.DefaultConfig()
	kwildConfig.DefaultConfig = func() *config.KwildConfig {
		cmdConfig.DefaultConfig(oldDefault)
		return oldDefault
	}

	root.AddCommand(kwildRoot.CustomRootCmd(cmdConfig.ProjectName, "node", cmdConfig.RootCmd))
	root.AddCommand(kwilCLIRoot.CustomRootCmd("client", cmdConfig.RootCmd, cmdConfig.ProjectName))
	root.AddCommand(kwilAdminRoot.CustomRootCmd("admin", cmdConfig.ProjectName))

	return root
}
