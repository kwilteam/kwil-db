package app

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/shared"
	"github.com/kwilteam/kwil-db/config"
)

// CmdConfig configures the root command.
type CmdConfig struct {
	// RootCmd is the name of the command.
	RootCmd string
	// ProjectName is the name of the project, which will be used in the help text.
	ProjectName string
	// DefaultConfig is a function that allows the default configuration to be changed.
	DefaultConfig func(*config.Config)
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

func NewCustomCmd(cmdConfig CmdConfig) *cobra.Command {
	custom.BinaryConfig.ProjectName = cmdConfig.ProjectName
	custom.BinaryConfig.ClientCmd = "client"
	custom.BinaryConfig.NodeCmd = "node"
	custom.BinaryConfig.RootCmd = cmdConfig.RootCmd

	root := &cobra.Command{
		Use:   cmdConfig.RootCmd,
		Short: "Command line interface client for using " + cmdConfig.ProjectName + ".",
		Long:  fmt.Sprintf(longDesc, cmdConfig.ProjectName, cmdConfig.ProjectName, cmdConfig.ProjectName),
	}

	if cmdConfig.DefaultConfig == nil {
		cmdConfig.DefaultConfig = func(c *config.Config) {}
	}

	if cmdConfig.ProjectName == "" {
		cmdConfig.ProjectName = "kwil"
	}

	if cmdConfig.RootCmd == "" {
		cmdConfig.RootCmd = cmdConfig.ProjectName
	}

	shared.DefaultConfig = func() *config.Config {
		initDefaultCfg := shared.DefaultConfig()
		cmdConfig.DefaultConfig(initDefaultCfg)
		return initDefaultCfg
	}

	root.AddCommand(RootCmd())
	// root.AddCommand(kwilCLIRoot.NewRootCmd()) // TODO

	return root
}
